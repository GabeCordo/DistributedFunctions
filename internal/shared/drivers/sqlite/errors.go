// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  errors.go
package sqlite

import (
	"database/sql"
	"errors"
)

var ErrNoRows = errors.New("sqlite: no rows in result set")

func IsNoRows(err error) bool {
	return errors.Is(err, ErrNoRows) || errors.Is(err, sql.ErrNoRows)
}
