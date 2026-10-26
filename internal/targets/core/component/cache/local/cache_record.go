// Package local
//
// Copyright (c) 2023-2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  cache_record.go
package local

import "time"

func (record Record) IsExpired() bool {
	// abstract how we understand cache expiry, at the moment cache
	// records expire at a fixed rate defined when the structure is
	// instantiated in minutes
	return time.Now().Sub(record.created).Minutes() >= record.expiry
}
