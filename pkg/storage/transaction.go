package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// memoryTransaction implements the Transaction interface
type memoryTransaction struct {
	tx     *sql.Tx
	engine *engineImpl
}

// Commit commits the transaction
func (mt *memoryTransaction) Commit(ctx context.Context) error {
	return mt.tx.Commit()
}

// Rollback rolls back the transaction
func (mt *memoryTransaction) Rollback(ctx context.Context) error {
	return mt.tx.Rollback()
}

// CreateTable creates a new table within the transaction
func (mt *memoryTransaction) CreateTable(name string, schema TableSchema) error {
	// For now, just defer to the engine
	// TODO: Implement proper transactional table creation
	return fmt.Errorf("transactional table creation not implemented")
}

// DropTable drops a table within the transaction
func (mt *memoryTransaction) DropTable(name string) error {
	// For now, just defer to the engine
	// TODO: Implement proper transactional table dropping
	return fmt.Errorf("transactional table dropping not implemented")
}

// CreateTensor creates a new tensor within the transaction
func (mt *memoryTransaction) CreateTensor(name string, schema TensorSchema) error {
	// For now, just defer to the engine
	// TODO: Implement proper transactional tensor creation
	return fmt.Errorf("transactional tensor creation not implemented")
}

// DropTensor drops a tensor within the transaction
func (mt *memoryTransaction) DropTensor(name string) error {
	// For now, just defer to the engine
	// TODO: Implement proper transactional tensor dropping
	return fmt.Errorf("transactional tensor dropping not implemented")
}
