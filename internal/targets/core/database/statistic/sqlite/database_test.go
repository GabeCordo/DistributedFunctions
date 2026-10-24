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
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic"
	"github.com/GabeCordo/ScalingFunctions"
)

////////////////////////////////////////////////////////////////////////
//						   Helper Functions
////////////////////////////////////////////////////////////////////////

// newTestSQLiteDatabase creates a SQLiteDatabase backed by a temporary file
// and closes it when the test finishes.
func newTestSQLiteDatabase(t *testing.T) *SQLiteDatabase {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_statistics.db")

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

// testStatisticsFilter returns the filter used by the CRUD tests.
func testStatisticsFilter() database.Filter {
	return database.Filter{Namespace: "test-namespace", Pipeline: "test-pipeline"}
}

// testStatisticsRecord returns a statistics record used by the CRUD tests.
func testStatisticsRecord() *ScalingFunctions.Statistics {
	return &ScalingFunctions.Statistics{
		NumOfFunctions: 2,
		Functions: []ScalingFunctions.FunctionStatistic{
			{Active: 1, Provisions: 10},
			{Active: 2, Provisions: 20},
		},
		NumOfChannels: 1,
		Pipes: []ScalingFunctions.PipeStatistic{
			{
				Pushed:   100,
				Pulled:   90,
				Dropped:  5,
				Breaches: 1,
				Timing: ScalingFunctions.TimingStatistics{
					MinTimeBeforePop: 1000,
					MaxTimeBeforePop: 2000,
					AverageTime:      1500,
					MedianTime:       1400,
				},
			},
		},
	}
}

////////////////////////////////////////////////////////////////////////
//						   		Tests
////////////////////////////////////////////////////////////////////////

func Test_SQLiteStatisticDatabase_Is_Implemented(t *testing.T) {
	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_statistics.db")

	d, err := NewSQLiteDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite database: %v", err)
	}

	var i statistic.Database
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

func Test_SQLiteStatisticDatabase_Get_Missing_Record(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	records := d.Get(database.Filter{Namespace: "does-not-exist"})
	if len(records) != 0 {
		t.Fatalf("Expected no statistics for a filter that matches nothing, got %d", len(records))
	}
}

func Test_SQLiteStatisticDatabase_Create_Then_Get(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testStatisticsRecord()
	filter := testStatisticsFilter()

	created, err := d.Create(filter, record)
	if err != nil {
		t.Fatalf("Failed to create statistic: %v", err)
	}
	if created == nil {
		t.Fatal("Expected created statistic to be non-nil")
	}

	records := d.Get(filter)
	if len(records) != 1 {
		t.Fatalf("Expected 1 statistic after create, got %d", len(records))
	}

	actual := records[0]
	if actual.NumOfFunctions != record.NumOfFunctions {
		t.Errorf("Expected num_of_functions %d, got %d", record.NumOfFunctions, actual.NumOfFunctions)
	}
	if actual.NumOfChannels != record.NumOfChannels {
		t.Errorf("Expected num_of_channels %d, got %d", record.NumOfChannels, actual.NumOfChannels)
	}
	if len(actual.Functions) != len(record.Functions) {
		t.Fatalf("Expected %d function statistics, got %d", len(record.Functions), len(actual.Functions))
	}
	for idx, expected := range record.Functions {
		fn := actual.Functions[idx]
		if fn.Active != expected.Active {
			t.Errorf("Expected function %d active %d, got %d", idx, expected.Active, fn.Active)
		}
		if fn.Provisions != expected.Provisions {
			t.Errorf("Expected function %d provisions %d, got %d", idx, expected.Provisions, fn.Provisions)
		}
	}
	if len(actual.Pipes) != len(record.Pipes) {
		t.Fatalf("Expected %d pipe statistics, got %d", len(record.Pipes), len(actual.Pipes))
	}
	for idx, expected := range record.Pipes {
		pipe := actual.Pipes[idx]
		if pipe.Pushed != expected.Pushed {
			t.Errorf("Expected pipe %d pushed %d, got %d", idx, expected.Pushed, pipe.Pushed)
		}
		if pipe.Pulled != expected.Pulled {
			t.Errorf("Expected pipe %d pulled %d, got %d", idx, expected.Pulled, pipe.Pulled)
		}
		if pipe.Dropped != expected.Dropped {
			t.Errorf("Expected pipe %d dropped %d, got %d", idx, expected.Dropped, pipe.Dropped)
		}
		if pipe.Breaches != expected.Breaches {
			t.Errorf("Expected pipe %d breaches %d, got %d", idx, expected.Breaches, pipe.Breaches)
		}
		if pipe.Timing != expected.Timing {
			t.Errorf("Expected pipe %d timing %+v, got %+v", idx, expected.Timing, pipe.Timing)
		}
	}
}

func Test_SQLiteStatisticDatabase_Create_Duplicate_Allowed(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testStatisticsRecord()
	filter := testStatisticsFilter()

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create statistic: %v", err)
	}

	created, err := d.Create(filter, record)
	if err != nil {
		t.Fatalf("Expected duplicate statistic creation to be allowed, got %v", err)
	}
	if created == nil {
		t.Fatal("Expected created statistic to be non-nil")
	}

	if records := d.Get(filter); len(records) != 2 {
		t.Fatalf("Expected 2 statistics after duplicate create, got %d", len(records))
	}
}

func Test_SQLiteStatisticDatabase_Create_Then_Delete(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testStatisticsRecord()
	filter := testStatisticsFilter()

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create statistic: %v", err)
	}

	if err := d.Delete(filter); err != nil {
		t.Fatalf("Expected delete of existing statistic to succeed, got %v", err)
	}

	if records := d.Get(filter); len(records) != 0 {
		t.Fatalf("Expected 0 statistics after delete, got %d", len(records))
	}
}

func Test_SQLiteStatisticDatabase_Delete_Missing_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	err := d.Delete(database.Filter{Namespace: "does-not-exist"})
	if err == nil {
		t.Fatal("Expected delete of a missing statistic to fail, got nil error")
	}
}

func Test_SQLiteStatisticDatabase_Replace_NotImplemented(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testStatisticsRecord()
	filter := testStatisticsFilter()

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create statistic: %v", err)
	}

	err := d.Replace(filter, testStatisticsRecord())
	if !errors.Is(err, database.NotImplemented) {
		t.Fatalf("Expected database.NotImplemented error, got %v", err)
	}
}

func Test_SQLiteStatisticDatabase_Replace_Missing_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testStatisticsRecord()

	err := d.Replace(database.Filter{Namespace: "does-not-exist", Pipeline: "does-not-exist"}, record)
	if !errors.Is(err, database.NotImplemented) {
		t.Fatalf("Expected database.NotImplemented error, got %v", err)
	}
}
