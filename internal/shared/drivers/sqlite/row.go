// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  row.go
package sqlite

import "database/sql"

// Row wraps sql.Row to hide the third-party database/sql dependency.
type Row struct {
	row *sql.Row
}

func (r Row) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}
