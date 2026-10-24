// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  rows.go
package sqlite

import "database/sql"

// Rows wraps sql.Rows to hide the third-party database/sql dependency.
type Rows struct {
	rows *sql.Rows
}

// Next prepares the next result row for reading.
// Returns true if there is a next row.
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Scan copies the columns from the current row into the dest arguments.
func (r *Rows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

// Close closes the rows iterator.
func (r *Rows) Close() error {
	return r.rows.Close()
}
