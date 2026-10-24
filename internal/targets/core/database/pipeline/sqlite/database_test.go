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
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/pipeline"
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
	dbPath := filepath.Join(tmpDir, "test_pipelines.db")

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

// testPipelineRecord returns a pipeline record used by the CRUD tests.
func testPipelineRecord(identifier string) *ScalingFunctions.PipelineIR {
	return &ScalingFunctions.PipelineIR{
		Identifier: identifier,
		OnCrash:    ScalingFunctions.Restart,
		OnStartup:  "on_startup_script",
		OnTeardown: "on_teardown_script",
		Functions: []ScalingFunctions.FunctionIR{
			{
				Module:     "test-module",
				Identifier: "function-1",
				From:       "pipe-in",
				To:         "pipe-out",
				StartWith:  2,
				WaitBefore: true,
				Maximum:    4,
				Parameters: []string{"param-1"},
				Returns:    []string{"return-1"},
				Metadata:   ScalingFunctions.FunctionMetadataIR{StaticMount: true},
			},
		},
		Pipes: []ScalingFunctions.PipeIR{
			{Identifier: "pipe-in", Threshold: 10, GrowthFactor: 1.5},
			{Identifier: "pipe-out", Threshold: 20, GrowthFactor: 2.5},
		},
	}
}

func testPipelineRecord2(identifier string) *ScalingFunctions.PipelineIR {
	return &ScalingFunctions.PipelineIR{
		Identifier: identifier,
		OnCrash:    ScalingFunctions.DoNothing,
		OnStartup:  "new_on_startup_script",
		OnTeardown: "new_on_teardown_script",
		Functions: []ScalingFunctions.FunctionIR{
			{
				Module:     "new-test-module",
				Identifier: "function-2",
				From:       "new-pipe-in",
				To:         "new-pipe-out",
				StartWith:  3,
				WaitBefore: false,
				Maximum:    6,
				Parameters: []string{"param-2"},
				Returns:    []string{"return-2"},
				Metadata:   ScalingFunctions.FunctionMetadataIR{StaticMount: false},
			},
		},
		Pipes: []ScalingFunctions.PipeIR{
			{Identifier: "new-pipe-in", Threshold: 30, GrowthFactor: 3.5},
		},
	}
}

////////////////////////////////////////////////////////////////////////
//						   		Tests
////////////////////////////////////////////////////////////////////////

func Test_SQLitePipelineDatabase_Is_Implemented(t *testing.T) {
	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_pipelines.db")

	d, err := NewSQLiteDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite database: %v", err)
	}

	var i pipeline.Database
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

func Test_SQLitePipelineDatabase_Get_Missing_Record(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	pipelines := d.Get(database.Filter{Namespace: "does-not-exist"})
	if len(pipelines) != 0 {
		t.Fatalf("Expected no pipelines for a filter that matches nothing, got %d", len(pipelines))
	}
}

func Test_SQLitePipelineDatabase_Create_Then_Get(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testPipelineRecord("pipeline-1")
	filter := database.Filter{Namespace: "test-namespace", Identifier: record.Identifier}

	identifier, err := d.Create(filter, record)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}
	if identifier != record.Identifier {
		t.Fatalf("Expected created pipeline identifier %q, got %q", record.Identifier, identifier)
	}

	pipelines := d.Get(database.Filter{Namespace: filter.Namespace})
	if len(pipelines) != 1 {
		t.Fatalf("Expected 1 pipeline after create, got %d", len(pipelines))
	}

	actual := pipelines[0]
	if actual.Identifier != record.Identifier {
		t.Errorf("Expected identifier %q, got %q", record.Identifier, actual.Identifier)
	}
	if actual.OnCrash != record.OnCrash {
		t.Errorf("Expected on_crash %q, got %q", record.OnCrash, actual.OnCrash)
	}
	if actual.OnStartup != record.OnStartup {
		t.Errorf("Expected on_startup %q, got %q", record.OnStartup, actual.OnStartup)
	}
	if actual.OnTeardown != record.OnTeardown {
		t.Errorf("Expected on_teardown %q, got %q", record.OnTeardown, actual.OnTeardown)
	}
	if len(actual.Functions) != len(record.Functions) {
		t.Fatalf("Expected %d functions, got %d", len(record.Functions), len(actual.Functions))
	}
	for idx, expected := range record.Functions {
		fn := actual.Functions[idx]
		if fn.Module != expected.Module {
			t.Errorf("Expected function %d module %q, got %q", idx, expected.Module, fn.Module)
		}
		if fn.Identifier != expected.Identifier {
			t.Errorf("Expected function %d identifier %q, got %q", idx, expected.Identifier, fn.Identifier)
		}
		if fn.From != expected.From {
			t.Errorf("Expected function %d from %q, got %q", idx, expected.From, fn.From)
		}
		if fn.To != expected.To {
			t.Errorf("Expected function %d to %q, got %q", idx, expected.To, fn.To)
		}
		if fn.StartWith != expected.StartWith {
			t.Errorf("Expected function %d start_with %d, got %d", idx, expected.StartWith, fn.StartWith)
		}
		if fn.WaitBefore != expected.WaitBefore {
			t.Errorf("Expected function %d wait_before %t, got %t", idx, expected.WaitBefore, fn.WaitBefore)
		}
		if fn.Maximum != expected.Maximum {
			t.Errorf("Expected function %d maximum %d, got %d", idx, expected.Maximum, fn.Maximum)
		}
		if fn.Metadata.StaticMount != expected.Metadata.StaticMount {
			t.Errorf("Expected function %d static_mount %t, got %t", idx, expected.Metadata.StaticMount, fn.Metadata.StaticMount)
		}
		if len(fn.Parameters) != len(expected.Parameters) {
			t.Fatalf("Expected function %d to have %d parameters, got %d", idx, len(expected.Parameters), len(fn.Parameters))
		}
		for pIdx, parameter := range expected.Parameters {
			if fn.Parameters[pIdx] != parameter {
				t.Errorf("Expected function %d parameter %q, got %q", idx, parameter, fn.Parameters[pIdx])
			}
		}
		if len(fn.Returns) != len(expected.Returns) {
			t.Fatalf("Expected function %d to have %d returns, got %d", idx, len(expected.Returns), len(fn.Returns))
		}
		for rIdx, returnValue := range expected.Returns {
			if fn.Returns[rIdx] != returnValue {
				t.Errorf("Expected function %d return %q, got %q", idx, returnValue, fn.Returns[rIdx])
			}
		}
	}
	if len(actual.Pipes) != len(record.Pipes) {
		t.Fatalf("Expected %d pipes, got %d", len(record.Pipes), len(actual.Pipes))
	}
	for idx, expected := range record.Pipes {
		pipe := actual.Pipes[idx]
		if pipe.Identifier != expected.Identifier {
			t.Errorf("Expected pipe %d identifier %q, got %q", idx, expected.Identifier, pipe.Identifier)
		}
		if pipe.Threshold != expected.Threshold {
			t.Errorf("Expected pipe %d threshold %d, got %d", idx, expected.Threshold, pipe.Threshold)
		}
		if pipe.GrowthFactor != expected.GrowthFactor {
			t.Errorf("Expected pipe %d growth_factor %f, got %f", idx, expected.GrowthFactor, pipe.GrowthFactor)
		}
	}
}

func Test_SQLitePipelineDatabase_Create_Duplicate_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testPipelineRecord("pipeline-1")
	filter := database.Filter{Namespace: "test-namespace", Identifier: record.Identifier}

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	_, err := d.Create(filter, record)
	if err == nil {
		t.Fatal("Expected duplicate pipeline creation to fail, got nil error")
	}
	if !errors.Is(err, database.AlreadyExists) {
		t.Fatalf("Expected database.AlreadyExists error, got %v", err)
	}
}

func Test_SQLitePipelineDatabase_Create_Then_Delete(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testPipelineRecord("pipeline-1")
	filter := database.Filter{Namespace: "test-namespace", Identifier: record.Identifier}

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := d.Delete(filter); err != nil {
		t.Fatalf("Expected delete of existing pipeline to succeed, got %v", err)
	}

	if pipelines := d.Get(database.Filter{Namespace: filter.Namespace}); len(pipelines) != 0 {
		t.Fatalf("Expected 0 pipelines after delete, got %d", len(pipelines))
	}
}

func Test_SQLitePipelineDatabase_Delete_Missing_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	err := d.Delete(database.Filter{Namespace: "does-not-exist", Identifier: "does-not-exist"})
	if err == nil {
		t.Fatal("Expected delete of a missing pipeline to fail, got nil error")
	}
}

func Test_SQLitePipelineDatabase_Create_Then_Replace(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testPipelineRecord("pipeline-1")
	filter := database.Filter{Namespace: "test-namespace", Identifier: record.Identifier}

	if _, err := d.Create(filter, record); err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	replacement := testPipelineRecord2(record.Identifier)

	if err := d.Replace(filter, replacement); err != nil {
		t.Fatalf("Expected replace of existing pipeline to succeed, got %v", err)
	}

	pipelines := d.Get(database.Filter{Namespace: filter.Namespace})
	if len(pipelines) != 1 {
		t.Fatalf("Expected 1 pipeline after replace, got %d", len(pipelines))
	}

	actual := pipelines[0]
	if actual.Identifier != replacement.Identifier {
		t.Errorf("Expected identifier %q, got %q", replacement.Identifier, actual.Identifier)
	}
	if actual.OnCrash != replacement.OnCrash {
		t.Errorf("Expected on_crash %q, got %q", replacement.OnCrash, actual.OnCrash)
	}
	if actual.OnStartup != replacement.OnStartup {
		t.Errorf("Expected on_startup %q, got %q", replacement.OnStartup, actual.OnStartup)
	}
	if actual.OnTeardown != replacement.OnTeardown {
		t.Errorf("Expected on_teardown %q, got %q", replacement.OnTeardown, actual.OnTeardown)
	}
	if len(actual.Functions) != len(replacement.Functions) {
		t.Fatalf("Expected %d functions after replace, got %d", len(replacement.Functions), len(actual.Functions))
	}
	fn := actual.Functions[0]
	expectedFn := replacement.Functions[0]
	if fn.Module != expectedFn.Module {
		t.Errorf("Expected function module %q, got %q", expectedFn.Module, fn.Module)
	}
	if fn.Identifier != expectedFn.Identifier {
		t.Errorf("Expected function identifier %q, got %q", expectedFn.Identifier, fn.Identifier)
	}
	if fn.From != expectedFn.From {
		t.Errorf("Expected function from %q, got %q", expectedFn.From, fn.From)
	}
	if fn.To != expectedFn.To {
		t.Errorf("Expected function to %q, got %q", expectedFn.To, fn.To)
	}
	if fn.StartWith != expectedFn.StartWith {
		t.Errorf("Expected function start_with %d, got %d", expectedFn.StartWith, fn.StartWith)
	}
	if fn.WaitBefore != expectedFn.WaitBefore {
		t.Errorf("Expected function wait_before %t, got %t", expectedFn.WaitBefore, fn.WaitBefore)
	}
	if fn.Maximum != expectedFn.Maximum {
		t.Errorf("Expected function maximum %d, got %d", expectedFn.Maximum, fn.Maximum)
	}
	if fn.Metadata.StaticMount != expectedFn.Metadata.StaticMount {
		t.Errorf("Expected function static_mount %t, got %t", expectedFn.Metadata.StaticMount, fn.Metadata.StaticMount)
	}
	if len(actual.Pipes) != len(replacement.Pipes) {
		t.Fatalf("Expected %d pipes after replace, got %d", len(replacement.Pipes), len(actual.Pipes))
	}
	pipe := actual.Pipes[0]
	expectedPipe := replacement.Pipes[0]
	if pipe.Identifier != expectedPipe.Identifier {
		t.Errorf("Expected pipe identifier %q, got %q", expectedPipe.Identifier, pipe.Identifier)
	}
	if pipe.Threshold != expectedPipe.Threshold {
		t.Errorf("Expected pipe threshold %d, got %d", expectedPipe.Threshold, pipe.Threshold)
	}
	if pipe.GrowthFactor != expectedPipe.GrowthFactor {
		t.Errorf("Expected pipe growth_factor %f, got %f", expectedPipe.GrowthFactor, pipe.GrowthFactor)
	}
}

func Test_SQLitePipelineDatabase_Replace_Missing_Fails(t *testing.T) {
	d := newTestSQLiteDatabase(t)

	record := testPipelineRecord("pipeline-1")

	err := d.Replace(database.Filter{Namespace: "does-not-exist", Identifier: "does-not-exist"}, record)
	if err == nil {
		t.Fatal("Expected replace of a missing pipeline to fail, got nil error")
	}
}
