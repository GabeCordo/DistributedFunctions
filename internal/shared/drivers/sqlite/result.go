// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  result.go
package sqlite

import "database/sql"

// Result wraps sql.Result to hide the third-party database/sql dependency.
type Result struct {
	result sql.Result
}

func (r Result) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}
