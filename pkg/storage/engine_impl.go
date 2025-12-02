package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/telumdb/telumdb/internal/config"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

// engineImpl implements the Engine interface
type engineImpl struct {
	config     *config.Config
	db         *sql.DB
	logger     *zap.Logger
	dataDir    string
	tensors    map[string]*tensorImpl
	tensorLock sync.RWMutex
	started    bool
}

// NewEngine creates a new storage engine instance
func NewEngine(cfg *config.Config) (Engine, error) {
	if err := os.MkdirAll(cfg.Storage.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(cfg.Storage.DataDir, "telumdb.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	engine := &engineImpl{
		config:  cfg,
		db:      db,
		dataDir: cfg.Storage.DataDir,
		tensors: make(map[string]*tensorImpl),
	}

	return engine, nil
}

// Start initializes the storage engine
func (e *engineImpl) Start(ctx context.Context) error {
	if e.started {
		return nil
	}

	// Initialize database schema
	if err := e.initSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Load existing tensors
	if err := e.loadTensors(); err != nil {
		return fmt.Errorf("failed to load tensors: %w", err)
	}

	e.started = true
	return nil
}

// Shutdown gracefully shuts down the storage engine
func (e *engineImpl) Shutdown(ctx context.Context) error {
	if !e.started {
		return nil
	}

	// Save all tensors
	e.tensorLock.Lock()
	for name, tensor := range e.tensors {
		if err := tensor.save(); err != nil {
			e.logger.Error("Failed to save tensor", zap.String("name", name), zap.Error(err))
		}
	}
	e.tensorLock.Unlock()

	// Close database
	if err := e.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	e.started = false
	return nil
}

// initSchema creates the necessary database tables
func (e *engineImpl) initSchema() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS telumdb_schema (
			version TEXT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tables (
			name TEXT PRIMARY KEY,
			schema TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS table_data (
			table_name TEXT NOT NULL,
			row_id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (table_name, row_id),
			FOREIGN KEY (table_name) REFERENCES tables(name) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS indexes (
			name TEXT PRIMARY KEY,
			table_name TEXT NOT NULL,
			columns TEXT NOT NULL,
			type TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (table_name) REFERENCES tables(name) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS tensors (
			name TEXT PRIMARY KEY,
			schema TEXT NOT NULL,
			metadata TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		if _, err := e.db.Exec(schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	// Insert schema version
	_, err := e.db.Exec(`INSERT OR REPLACE INTO telumdb_schema (version) VALUES (?)`, "1.0")
	if err != nil {
		return fmt.Errorf("failed to set schema version: %w", err)
	}

	return nil
}

// CreateTable creates a new table
func (e *engineImpl) CreateTable(name string, schema TableSchema) error {
	if !e.started {
		return fmt.Errorf("engine not started")
	}

	// Serialize schema
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to serialize schema: %w", err)
	}

	// Insert table metadata
	_, err = e.db.Exec(
		`INSERT INTO tables (name, schema) VALUES (?, ?)`,
		name, string(schemaJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// DropTable removes a table
func (e *engineImpl) DropTable(name string) error {
	if !e.started {
		return fmt.Errorf("engine not started")
	}

	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete table data
	_, err = tx.Exec(`DELETE FROM table_data WHERE table_name = ?`, name)
	if err != nil {
		return fmt.Errorf("failed to delete table data: %w", err)
	}

	// Delete indexes
	_, err = tx.Exec(`DELETE FROM indexes WHERE table_name = ?`, name)
	if err != nil {
		return fmt.Errorf("failed to delete indexes: %w", err)
	}

	// Delete table
	_, err = tx.Exec(`DELETE FROM tables WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}

	return tx.Commit()
}

// GetTable retrieves a table
func (e *engineImpl) GetTable(name string) (Table, error) {
	if !e.started {
		return nil, fmt.Errorf("engine not started")
	}

	var schemaJSON string
	err := e.db.QueryRow(
		`SELECT schema FROM tables WHERE name = ?`,
		name,
	).Scan(&schemaJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("table not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get table: %w", err)
	}

	var schema TableSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("failed to deserialize schema: %w", err)
	}

	return &memoryTable{
		name:   name,
		schema: schema,
		engine: e,
	}, nil
}

// ListTables returns all table names
func (e *engineImpl) ListTables() ([]string, error) {
	if !e.started {
		return nil, fmt.Errorf("engine not started")
	}

	rows, err := e.db.Query(`SELECT name FROM tables ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, name)
	}

	return tables, nil
}

// CreateTensor creates a new tensor
func (e *engineImpl) CreateTensor(name string, schema TensorSchema) error {
	if !e.started {
		return fmt.Errorf("engine not started")
	}

	e.tensorLock.Lock()
	defer e.tensorLock.Unlock()

	// Check if tensor already exists
	if _, exists := e.tensors[name]; exists {
		return fmt.Errorf("tensor already exists: %s", name)
	}

	// Serialize schema
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to serialize tensor schema: %w", err)
	}

	// Insert tensor metadata
	_, err = e.db.Exec(
		`INSERT INTO tensors (name, schema, metadata) VALUES (?, ?, ?)`,
		name, string(schemaJSON), "{}",
	)
	if err != nil {
		return fmt.Errorf("failed to create tensor: %w", err)
	}

	// Create tensor instance
	tensor := &tensorImpl{
		name:   name,
		schema: schema,
		engine: e,
		data:   make([]float32, e.calculateTensorSize(schema)),
	}

	e.tensors[name] = tensor

	// Save tensor data
	if err := tensor.save(); err != nil {
		delete(e.tensors, name)
		return fmt.Errorf("failed to save tensor: %w", err)
	}

	return nil
}

// DropTensor removes a tensor
func (e *engineImpl) DropTensor(name string) error {
	if !e.started {
		return fmt.Errorf("engine not started")
	}

	e.tensorLock.Lock()
	defer e.tensorLock.Unlock()

	// Remove from memory
	if tensor, exists := e.tensors[name]; exists {
		tensorPath := tensor.getFilePath()
		os.Remove(tensorPath)
		delete(e.tensors, name)
	}

	// Remove from database
	_, err := e.db.Exec(`DELETE FROM tensors WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("failed to delete tensor: %w", err)
	}

	return nil
}

// GetTensor retrieves a tensor
func (e *engineImpl) GetTensor(name string) (Tensor, error) {
	if !e.started {
		return nil, fmt.Errorf("engine not started")
	}

	e.tensorLock.RLock()
	tensor, exists := e.tensors[name]
	e.tensorLock.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tensor not found: %s", name)
	}

	return tensor, nil
}

// ListTensors returns all tensor names
func (e *engineImpl) ListTensors() ([]string, error) {
	if !e.started {
		return nil, fmt.Errorf("engine not started")
	}

	rows, err := e.db.Query(`SELECT name FROM tensors ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tensors: %w", err)
	}
	defer rows.Close()

	var tensors []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan tensor name: %w", err)
		}
		tensors = append(tensors, name)
	}

	return tensors, nil
}

// ExecuteQuery executes a query and returns results
func (e *engineImpl) ExecuteQuery(ctx context.Context, query string) (Result, error) {
	if !e.started {
		return Result{}, fmt.Errorf("engine not started")
	}

	// For now, implement basic SQL execution
	// TODO: Add TQL parsing and execution
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return Result{}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return Result{}, fmt.Errorf("failed to get columns: %w", err)
	}

	// Read all rows
	var rowData [][]interface{}
	for rows.Next() {
		// Create slice for row values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			return Result{}, fmt.Errorf("failed to scan row: %w", err)
		}

		rowData = append(rowData, values)
	}

	// Get affected rows count (for SELECT, this is typically 0)
	affected := int64(0)

	return Result{
		Columns:  columns,
		Rows:     rowData,
		Affected: affected,
	}, nil
}

// BeginTransaction starts a new transaction
func (e *engineImpl) BeginTransaction(ctx context.Context) (Transaction, error) {
	if !e.started {
		return nil, fmt.Errorf("engine not started")
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &memoryTransaction{
		tx:     tx,
		engine: e,
	}, nil
}

// Helper methods

func (e *engineImpl) calculateTensorSize(schema TensorSchema) int {
	size := 1
	for _, dim := range schema.Shape {
		size *= dim
	}
	return size
}

func (e *engineImpl) loadTensors() error {
	e.tensorLock.Lock()
	defer e.tensorLock.Unlock()

	rows, err := e.db.Query(`SELECT name, schema, metadata FROM tensors`)
	if err != nil {
		return fmt.Errorf("failed to load tensors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, schemaJSON, metadataJSON string
		if err := rows.Scan(&name, &schemaJSON, &metadataJSON); err != nil {
			return fmt.Errorf("failed to scan tensor: %w", err)
		}

		var schema TensorSchema
		if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
			return fmt.Errorf("failed to deserialize tensor schema: %w", err)
		}

		tensor := &tensorImpl{
			name:   name,
			schema: schema,
			engine: e,
			data:   make([]float32, e.calculateTensorSize(schema)),
		}

		// Load tensor data from file
		if err := tensor.load(); err != nil {
			e.logger.Warn("Failed to load tensor data", zap.String("name", name), zap.Error(err))
		}

		e.tensors[name] = tensor
	}

	return nil
}
