package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// memoryTable implements the Table interface
type memoryTable struct {
	name   string
	schema TableSchema
	engine Engine
}

// Name returns the table name
func (t *memoryTable) Name() string {
	return t.name
}

// Schema returns the table schema
func (t *memoryTable) Schema() TableSchema {
	return t.schema
}

// Insert inserts a new row into the table
func (t *memoryTable) Insert(ctx context.Context, row Row) error {
	// Serialize row data
	rowJSON, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("failed to serialize row: %w", err)
	}

	// Generate row ID
	rowID := generateRowID()

	// Insert into database
	engine := t.engine.(*engineImpl)
	_, err = engine.db.Exec(
		`INSERT INTO table_data (table_name, row_id, data) VALUES (?, ?, ?)`,
		t.name, rowID, string(rowJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to insert row: %w", err)
	}

	return nil
}

// Update updates rows matching the condition
func (t *memoryTable) Update(ctx context.Context, row Row, condition Condition) error {
	// For now, implement simple update based on row ID
	// TODO: Implement proper condition parsing
	if id, ok := row["id"]; ok {
		rowJSON, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("failed to serialize row: %w", err)
		}

		engine := t.engine.(*engineImpl)
		_, err = engine.db.Exec(
			`UPDATE table_data SET data = ? WHERE table_name = ? AND row_id = ?`,
			string(rowJSON), t.name, fmt.Sprintf("%v", id),
		)
		if err != nil {
			return fmt.Errorf("failed to update row: %w", err)
		}
	}

	return nil
}

// Delete deletes rows matching the condition
func (t *memoryTable) Delete(ctx context.Context, condition Condition) error {
	// For now, implement simple delete based on ID condition
	// TODO: Implement proper condition parsing
	engine := t.engine.(*engineImpl)

	if condition == nil {
		// Delete all rows
		_, err := engine.db.Exec(
			`DELETE FROM table_data WHERE table_name = ?`,
			t.name,
		)
		if err != nil {
			return fmt.Errorf("failed to delete rows: %w", err)
		}
	} else {
		// Simple ID-based deletion using condition string
		conditionStr := condition.String()
		if strings.Contains(conditionStr, "id") {
			// Extract ID from condition string (simple parsing)
			parts := strings.Fields(conditionStr)
			if len(parts) >= 3 {
				id := parts[2]
				_, err := engine.db.Exec(
					`DELETE FROM table_data WHERE table_name = ? AND row_id = ?`,
					t.name, id,
				)
				if err != nil {
					return fmt.Errorf("failed to delete row: %w", err)
				}
			}
		}
	}

	return nil
}

// Select retrieves rows matching the condition
func (t *memoryTable) Select(ctx context.Context, columns []string, condition Condition) (Iterator, error) {
	// Build query
	query := fmt.Sprintf("SELECT row_id, data FROM table_data WHERE table_name = '%s'", t.name)

	if condition != nil {
		// Simple condition handling using condition string
		conditionStr := condition.String()
		if strings.Contains(conditionStr, "id") {
			// Extract ID from condition string (simple parsing)
			parts := strings.Fields(conditionStr)
			if len(parts) >= 3 {
				id := parts[2]
				query += fmt.Sprintf(" AND row_id = '%s'", id)
			}
		}
	}

	engine := t.engine.(*engineImpl)
	rows, err := engine.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to select rows: %w", err)
	}

	return &memoryIterator{
		rows:    rows,
		columns: columns,
		table:   t,
	}, nil
}

// Count returns the number of rows matching the condition
func (t *memoryTable) Count(ctx context.Context, condition Condition) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM table_data WHERE table_name = '%s'", t.name)

	if condition != nil {
		// Simple condition handling using condition string
		conditionStr := condition.String()
		if strings.Contains(conditionStr, "id") {
			// Extract ID from condition string (simple parsing)
			parts := strings.Fields(conditionStr)
			if len(parts) >= 3 {
				id := parts[2]
				query += fmt.Sprintf(" AND row_id = '%s'", id)
			}
		}
	}

	engine := t.engine.(*engineImpl)
	var count int64
	err := engine.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count rows: %w", err)
	}

	return count, nil
}

// memoryIterator implements the Iterator interface
type memoryIterator struct {
	rows    *sql.Rows
	columns []string
	table   *memoryTable
	closed  bool
}

// Next advances to the next row
func (it *memoryIterator) Next() bool {
	if it.closed {
		return false
	}

	hasNext := it.rows.Next()
	if !hasNext {
		it.Close()
	}
	return hasNext
}

// Scan copies the current row's values into the provided destinations
func (it *memoryIterator) Scan(dest ...interface{}) error {
	if it.closed {
		return fmt.Errorf("iterator is closed")
	}

	var rowID string
	var dataJSON string

	err := it.rows.Scan(&rowID, &dataJSON)
	if err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Parse row data
	var rowData map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &rowData); err != nil {
		return fmt.Errorf("failed to parse row data: %w", err)
	}

	// Map columns to destinations
	if len(it.columns) == 0 {
		// Return all columns
		if len(dest) > 0 {
			if destPtr, ok := dest[0].(*map[string]interface{}); ok {
				*destPtr = rowData
			}
		}
	} else {
		// Return specific columns
		for i, col := range it.columns {
			if i < len(dest) {
				if value, ok := rowData[col]; ok {
					switch v := dest[i].(type) {
					case *string:
						*v = fmt.Sprintf("%v", value)
					case *int:
						if intVal, ok := value.(float64); ok {
							*v = int(intVal)
						}
					case *int64:
						if intVal, ok := value.(float64); ok {
							*v = int64(intVal)
						}
					case *float64:
						if floatVal, ok := value.(float64); ok {
							*v = floatVal
						}
					case *bool:
						if boolVal, ok := value.(bool); ok {
							*v = boolVal
						}
					}
				}
			}
		}
	}

	return nil
}

// Close closes the iterator
func (it *memoryIterator) Close() error {
	if !it.closed {
		it.closed = true
		return it.rows.Close()
	}
	return nil
}

// Columns returns the column names
func (it *memoryIterator) Columns() []string {
	return it.columns
}

// Helper functions

func generateRowID() string {
	return fmt.Sprintf("row_%d", len("placeholder")) // Simple ID generation
}
