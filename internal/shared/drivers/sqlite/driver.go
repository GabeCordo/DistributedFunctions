// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  driver.go
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Driver represents a connection to a sqlite database.
//
// Developers should use a `Driver` to maintain avoid tight coupling to the third-paty
// sqlite dependency used to facilitate interactions with the sqlite file.
type Driver struct {
	db *sql.DB
}

// NewDriver creates a new `Driver` instance.
func NewDriver(path string) (*Driver, error) {
	d := new(Driver)

	dir := filepath.Dir(path)

	// The `path` shall be an absolute path to a file.
	if dir == "" || dir == "." {
		return nil, fmt.Errorf("path must be an absolute path")
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	d.db = db

	err = d.db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return d, nil
}

// IsConnected checks if the driver is connected to the database
func (d *Driver) IsConnected() bool {
	if d.db == nil {
		return false
	}

	return d.db.Ping() == nil
}

// Close closes the database connection
func (d *Driver) Close() error {
	if d.db != nil {
		return d.db.Close()
	} else {
		return nil
	}
}

// Exec executes a query on the sqlite database without returning rows.
func (d *Driver) Exec(query string, args ...any) error {
	_, err := d.db.Exec(query, args...)
	return err
}

// ExecResult executes a query that doesn't return rows and returns a Result.
func (d *Driver) ExecResult(query string, args ...any) (Result, error) {
	result, err := d.db.Exec(query, args...)
	return Result{result: result}, err
}

// Query executes a query that returns rows
func (d *Driver) Query(query string, args ...any) (*Rows, error) {
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// QueryRow executes a query that returns at most one row
func (d *Driver) QueryRow(query string, args ...any) Row {
	return Row{row: d.db.QueryRow(query, args...)}
}

// BeginTransaction starts a transaction and returns a wrapped Transaction.
// This is the preferred method for starting transactions.
func (d *Driver) BeginTransaction() (Transaction, error) {
	tx, err := d.db.Begin()
	return Transaction{tx: tx}, err
}
