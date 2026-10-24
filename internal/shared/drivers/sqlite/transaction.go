// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  transaction.go
package sqlite

import "database/sql"

// Transaction wraps sql.Tx to provide a clean interface for database transactions.
// It holds a single pointer to the underlying sql.Tx, so passing by value is safe and idiomatic.
type Transaction struct {
	tx *sql.Tx
}

// Exec executes a query that doesn't return rows within the transaction.
func (t Transaction) Exec(query string, args ...any) (Result, error) {
	result, err := t.tx.Exec(query, args...)
	return Result{result: result}, err
}

// Rollback aborts the transaction.
func (t Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Commit commits the transaction.
func (t Transaction) Commit() error {
	return t.tx.Commit()
}
