// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  database_test.go
package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

func Test_SQLiteJobDatabase_Is_Implemented(t *testing.T) {
	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_jobs.db")

	d, err := NewSQLiteDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite database: %v", err)
	}

	var i job.Database
	i = d

	i.Print()

	// Clean up
	if err := d.Close(); err != nil {
		t.Logf("Warning: failed to close database: %v", err)
	}
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: failed to remove temporary database file: %v", err)
	}
}
