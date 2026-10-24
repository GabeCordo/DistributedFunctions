// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  database_test.go
package sqlite

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

////////////////////////////////////////////////////////////////////////
//						   Helper Functions
////////////////////////////////////////////////////////////////////////

// newTestSQLiteDatabase creates a SQLiteDatabase backed by a temporary file
// and closes it when the test finishes.
func newTestSQLiteDatabase(t *testing.T) *SQLiteDatabase {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_jobs.db")

	d, err := NewSQLiteDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite database: %v", err)
	}

	if err := d.Load(dbPath); err != nil {
		t.Fatalf("Failed to load SQLite database: %v", err)
	}

	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	})

	return d
}

// testJobRecord returns a job record used by the CRUD tests.
func testJobRecord(identifier string) *job.Job {
	return &job.Job{
		Identifier: identifier,
		Namespace:  "test-namespace",
		Pipeline:   "test-pipeline",
		Interval: database.Interval{
			Minute: 5,
			Hour:   1,
			Day:    2,
			Month:  3,
		},
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
}

func testJobRecord2(identifier string) *job.Job {
	return &job.Job{
		Identifier: identifier,
		Namespace:  "test-namespace",
		Pipeline:   "test-pipeline",
		Interval: database.Interval{
			Minute: 10,
			Hour:   2,
			Day:    4,
			Month:  6,
		},
		Metadata: map[string]string{
			"key3": "value3",
		},
	}
}

////////////////////////////////////////////////////////////////////////
//						   		Tests
////////////////////////////////////////////////////////////////////////

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

func Test_SQLiteJobDatabase_Get_Missing_Record(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	jobs := d.Get(database.Filter{Identifier: "does-not-exist"})
	if len(jobs) != 0 {
		t.Fatalf("Expected no jobs for a filter that matches nothing, got %d", len(jobs))
	}
}

func Test_SQLiteJobDatabase_Create_Then_Get(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testJobRecord("job-1")
	filter := database.Filter{Identifier: record.Identifier}

	identifier, err := d.Create(filter, record)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}
	if identifier != record.Identifier {
		t.Fatalf("Expected created job identifier %q, got %q", record.Identifier, identifier)
	}

	jobs := d.Get(filter)
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job after create, got %d", len(jobs))
	}

	actual := jobs[0]
	if actual.Identifier != record.Identifier {
		t.Errorf("Expected identifier %q, got %q", record.Identifier, actual.Identifier)
	}
	if actual.Namespace != record.Namespace {
		t.Errorf("Expected namespace %q, got %q", record.Namespace, actual.Namespace)
	}
	if actual.Pipeline != record.Pipeline {
		t.Errorf("Expected pipeline %q, got %q", record.Pipeline, actual.Pipeline)
	}
	if actual.Interval != record.Interval {
		t.Errorf("Expected interval %+v, got %+v", record.Interval, actual.Interval)
	}
	if len(actual.Metadata) != len(record.Metadata) {
		t.Fatalf("Expected %d metadata entries, got %d", len(record.Metadata), len(actual.Metadata))
	}
	for key, value := range record.Metadata {
		if actual.Metadata[key] != value {
			t.Errorf("Expected metadata %q = %q, got %q", key, value, actual.Metadata[key])
		}
	}
}

func Test_SQLiteJobDatabase_Create_Duplicate_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testJobRecord("job-1")
	filter := database.Filter{Identifier: record.Identifier}

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	_, err := d.Create(filter, record)
	if err == nil {
		t.Fatal("Expected duplicate job creation to fail, got nil error")
	}
	if !errors.Is(err, database.AlreadyExists) {
		t.Fatalf("Expected database.AlreadyExists error, got %v", err)
	}
}

func Test_SQLiteJobDatabase_Create_Then_Delete(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testJobRecord("job-1")
	filter := database.Filter{Identifier: record.Identifier}

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if err := d.Delete(filter); err != nil {
		t.Fatalf("Expected delete of existing job to succeed, got %v", err)
	}

	if jobs := d.Get(filter); len(jobs) != 0 {
		t.Fatalf("Expected 0 jobs after delete, got %d", len(jobs))
	}
}

func Test_SQLiteJobDatabase_Delete_Missing_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	err := d.Delete(database.Filter{Identifier: "does-not-exist"})
	if err == nil {
		t.Fatal("Expected delete of a missing job to fail, got nil error")
	}
}

func Test_SQLiteJobDatabase_Create_Then_Replace(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testJobRecord("job-1")
	filter := database.Filter{Identifier: record.Identifier}

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	replacement := testJobRecord2("job-1-replaced")

	if err := d.Replace(filter, replacement); err != nil {
		t.Fatalf("Expected replace of existing job to succeed, got %v", err)
	}

	// the original identifier must no longer match any record
	if jobs := d.Get(filter); len(jobs) != 0 {
		t.Fatalf("Expected 0 jobs under the original identifier after replace, got %d", len(jobs))
	}

	replacedJobs := d.Get(database.Filter{Identifier: replacement.Identifier})
	if len(replacedJobs) != 1 {
		t.Fatalf("Expected 1 job after replace, got %d", len(replacedJobs))
	}

	actual := replacedJobs[0]
	if actual.Identifier != replacement.Identifier {
		t.Errorf("Expected identifier %q, got %q", replacement.Identifier, actual.Identifier)
	}
	if actual.Namespace != replacement.Namespace {
		t.Errorf("Expected namespace %q, got %q", replacement.Namespace, actual.Namespace)
	}
	if actual.Pipeline != replacement.Pipeline {
		t.Errorf("Expected pipeline %q, got %q", replacement.Pipeline, actual.Pipeline)
	}
	if actual.Interval != replacement.Interval {
		t.Errorf("Expected interval %+v, got %+v", replacement.Interval, actual.Interval)
	}
	if len(actual.Metadata) != len(replacement.Metadata) {
		t.Fatalf("Expected %d metadata entries, got %d", len(replacement.Metadata), len(actual.Metadata))
	}
	for key, value := range replacement.Metadata {
		if actual.Metadata[key] != value {
			t.Errorf("Expected metadata %q = %q, got %q", key, value, actual.Metadata[key])
		}
	}
}

func Test_SQLiteJobDatabase_Replace_Missing_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testJobRecord("job-1")

	err := d.Replace(database.Filter{Identifier: "does-not-exist"}, record)
	if err == nil {
		t.Fatal("Expected replace of a missing job to fail, got nil error")
	}
}
