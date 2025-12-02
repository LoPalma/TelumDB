package storage

import (
	"context"
	"testing"

	"github.com/telumdb/telumdb/internal/config"
)

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name    string
		config  config.StorageConfig
		wantErr bool
	}{
		{
			name: "hybrid engine",
			config: config.StorageConfig{
				Engine: "hybrid",
			},
			wantErr: false,
		},
		{
			name: "memory engine",
			config: config.StorageConfig{
				Engine: "memory",
			},
			wantErr: false,
		},
		{
			name: "unsupported engine",
			config: config.StorageConfig{
				Engine: "unsupported",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateEngine(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryEngine(t *testing.T) {
	cfg := config.StorageConfig{
		Engine: "memory",
	}

	engine, err := CreateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create memory engine: %v", err)
	}

	ctx := context.Background()

	// Test start and shutdown
	if err := engine.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if err := engine.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestMemoryEngineTables(t *testing.T) {
	cfg := config.StorageConfig{
		Engine: "memory",
	}

	engine, err := CreateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create memory engine: %v", err)
	}

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Shutdown(ctx)

	// Test empty table list
	tables, err := engine.ListTables()
	if err != nil {
		t.Errorf("ListTables() error = %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("Expected 0 tables, got %d", len(tables))
	}

	// Test table operations
	tableName := "test_table"
	schema := TableSchema{
		Columns: []ColumnDefinition{
			{Name: "id", Type: "INTEGER", Nullable: false},
			{Name: "name", Type: "VARCHAR", Nullable: true},
		},
	}

	// Create table
	if err := engine.CreateTable(tableName, schema); err == nil {
		t.Error("CreateTable() should not be implemented yet")
	}

	// Get non-existent table
	_, err = engine.GetTable(tableName)
	if err == nil {
		t.Error("GetTable() should return error for non-existent table")
	}

	// Drop table
	if err := engine.DropTable(tableName); err != nil {
		t.Errorf("DropTable() error = %v", err)
	}
}

func TestMemoryEngineTensors(t *testing.T) {
	cfg := config.StorageConfig{
		Engine: "memory",
	}

	engine, err := CreateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create memory engine: %v", err)
	}

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Shutdown(ctx)

	// Test empty tensor list
	tensors, err := engine.ListTensors()
	if err != nil {
		t.Errorf("ListTensors() error = %v", err)
	}
	if len(tensors) != 0 {
		t.Errorf("Expected 0 tensors, got %d", len(tensors))
	}

	// Test tensor operations
	tensorName := "test_tensor"
	schema := TensorSchema{
		Shape:       []int{10, 20},
		DType:       "float32",
		ChunkSize:   []int{5, 5},
		Compression: "none",
		Metadata:    map[string]interface{}{"description": "test tensor"},
	}

	// Create tensor
	if err := engine.CreateTensor(tensorName, schema); err == nil {
		t.Error("CreateTensor() should not be implemented yet")
	}

	// Get non-existent tensor
	_, err = engine.GetTensor(tensorName)
	if err == nil {
		t.Error("GetTensor() should return error for non-existent tensor")
	}

	// Drop tensor
	if err := engine.DropTensor(tensorName); err != nil {
		t.Errorf("DropTensor() error = %v", err)
	}
}
