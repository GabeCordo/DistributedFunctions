// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  nullstring.go
package sqlite

import "database/sql"

// NullString wraps sql.NullString to hide the third-party database/sql dependency.
type NullString struct {
	nullString sql.NullString
}

func (n NullString) Valid() bool {
	return n.nullString.Valid
}

func (n NullString) String() string {
	return n.nullString.String
}

// NewNullString creates a new NullString with the given value and validity.
func NewNullString(value string, valid bool) NullString {
	return NullString{
		nullString: sql.NullString{
			String: value,
			Valid:  valid,
		},
	}
}
