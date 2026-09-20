// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved.
//
// Source file:  queryBuilder.go
package sqlite

////////////////////////////////////////////////////////////////////////
//					SQLiteDatabaseQueryBuilder
////////////////////////////////////////////////////////////////////////

// SQLiteDatabaseQueryBuilder builds SQL queries with pre-allocated buffers
type SQLiteDatabaseQueryBuilder struct {
	conditions []string
	args       []any
}

// newSQLiteDatabaseQueryBuilder creates a new query builder with pre-allocated slices
func newSQLiteDatabaseQueryBuilder() *SQLiteDatabaseQueryBuilder {
	return &SQLiteDatabaseQueryBuilder{
		conditions: make([]string, 0, 4),
		args:       make([]any, 0, 8),
	}
}

// getCreateTablesSQL returns the SQL to create all pipeline tables
func (b *SQLiteDatabaseQueryBuilder) getCreateTablesSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS pipeline (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		namespace TEXT NOT NULL,
		identifier TEXT NOT NULL,
		UNIQUE(namespace, identifier)
	);

	CREATE TABLE IF NOT EXISTS pipeline_ir (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pipeline_id INTEGER NOT NULL,
		on_crash TEXT,
		on_startup TEXT,
		on_teardown TEXT,
		FOREIGN KEY (pipeline_id) REFERENCES pipeline(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS function_ir (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pipeline_ir_id INTEGER NOT NULL,
		module TEXT NOT NULL,
		identifier TEXT NOT NULL,
		from_pipe TEXT,
		to_pipe TEXT,
		start_with INTEGER NOT NULL DEFAULT 1,
		wait_before INTEGER NOT NULL DEFAULT 0,
		maximum INTEGER NOT NULL DEFAULT 1,
		parameters TEXT NOT NULL DEFAULT '[]',
		returns TEXT NOT NULL DEFAULT '[]',
		static_mount INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (pipeline_ir_id) REFERENCES pipeline_ir(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS pipe_ir (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pipeline_ir_id INTEGER NOT NULL,
		identifier TEXT NOT NULL,
		threshold INTEGER NOT NULL DEFAULT 1,
		growth_factor REAL NOT NULL DEFAULT 2.0,
		FOREIGN KEY (pipeline_ir_id) REFERENCES pipeline_ir(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_pipeline_namespace ON pipeline(namespace);
	CREATE INDEX IF NOT EXISTS idx_pipeline_identifier ON pipeline(identifier);
	CREATE INDEX IF NOT EXISTS idx_pipeline_ir_pipeline_id ON pipeline_ir(pipeline_id);
	CREATE INDEX IF NOT EXISTS idx_function_ir_pipeline_ir_id ON function_ir(pipeline_ir_id);
	CREATE INDEX IF NOT EXISTS idx_pipe_ir_pipeline_ir_id ON pipe_ir(pipeline_ir_id);
	`
}

// buildGetPipelineQuery builds the SQL to get a pipeline by namespace and identifier
func (b *SQLiteDatabaseQueryBuilder) buildGetPipelineQuery(namespace, identifier string) (string, []any) {
	b.conditions = b.conditions[:0]
	b.args = b.args[:0]

	b.conditions = append(b.conditions, "p.namespace = ?")
	b.args = append(b.args, namespace)
	b.conditions = append(b.conditions, "p.identifier = ?")
	b.args = append(b.args, identifier)

	query := `SELECT p.id, p.namespace, p.identifier FROM pipeline p WHERE ` + b.conditions[0]
	for i := 1; i < len(b.conditions); i++ {
		query += " AND " + b.conditions[i]
	}
	return query, b.args
}

// buildGetPipelineIRQuery builds the SQL to get pipeline_ir by pipeline_id
func (b *SQLiteDatabaseQueryBuilder) buildGetPipelineIRQuery(pipelineID int64) (string, []any) {
	return `SELECT id, pipeline_id, on_crash, on_startup, on_teardown FROM pipeline_ir WHERE pipeline_id = ?`, []any{pipelineID}
}

// buildGetFunctionIRsQuery builds the SQL to get function_irs by pipeline_ir_id
func (b *SQLiteDatabaseQueryBuilder) buildGetFunctionIRsQuery(pipelineIRID int64) (string, []any) {
	return `SELECT id, pipeline_ir_id, module, identifier, from_pipe, to_pipe, start_with, wait_before, maximum, parameters, returns, static_mount FROM function_ir WHERE pipeline_ir_id = ? ORDER BY id`, []any{pipelineIRID}
}

// buildGetPipeIRsQuery builds the SQL to get pipe_irs by pipeline_ir_id
func (b *SQLiteDatabaseQueryBuilder) buildGetPipeIRsQuery(pipelineIRID int64) (string, []any) {
	return `SELECT id, pipeline_ir_id, identifier, threshold, growth_factor FROM pipe_ir WHERE pipeline_ir_id = ? ORDER BY id`, []any{pipelineIRID}
}

// buildInsertPipelineQuery builds the SQL to insert a pipeline
func (b *SQLiteDatabaseQueryBuilder) buildInsertPipelineQuery(namespace, identifier string) (string, []any) {
	return `INSERT INTO pipeline (namespace, identifier) VALUES (?, ?)`, []any{namespace, identifier}
}

// buildInsertPipelineIRQuery builds the SQL to insert a pipeline_ir
func (b *SQLiteDatabaseQueryBuilder) buildInsertPipelineIRQuery(pipelineID int64, onCrash, onStartup, onTeardown string) (string, []any) {
	return `INSERT INTO pipeline_ir (pipeline_id, on_crash, on_startup, on_teardown) VALUES (?, ?, ?, ?)`, []any{pipelineID, onCrash, onStartup, onTeardown}
}

// buildInsertFunctionIRQuery builds the SQL to insert a function_ir
func (b *SQLiteDatabaseQueryBuilder) buildInsertFunctionIRQuery(pipelineIRID int64, module, identifier, fromPipe, toPipe string, startWith, waitBefore, maximum int, parameters, returns string, staticMount bool) (string, []any) {
	staticMountInt := 0
	if staticMount {
		staticMountInt = 1
	}
	return `INSERT INTO function_ir (pipeline_ir_id, module, identifier, from_pipe, to_pipe, start_with, wait_before, maximum, parameters, returns, static_mount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		[]any{pipelineIRID, module, identifier, fromPipe, toPipe, startWith, waitBefore, maximum, parameters, returns, staticMountInt}
}

// buildInsertPipeIRQuery builds the SQL to insert a pipe_ir
func (b *SQLiteDatabaseQueryBuilder) buildInsertPipeIRQuery(pipelineIRID int64, identifier string, threshold int, growthFactor float64) (string, []any) {
	return `INSERT INTO pipe_ir (pipeline_ir_id, identifier, threshold, growth_factor) VALUES (?, ?, ?, ?)`, []any{pipelineIRID, identifier, threshold, growthFactor}
}

// buildDeleteFunctionIRsQuery builds the SQL to delete function_irs by pipeline_ir_id
func (b *SQLiteDatabaseQueryBuilder) buildDeleteFunctionIRsQuery(pipelineIRID int64) (string, []any) {
	return `DELETE FROM function_ir WHERE pipeline_ir_id = ?`, []any{pipelineIRID}
}

// buildDeletePipeIRsQuery builds the SQL to delete pipe_irs by pipeline_ir_id
func (b *SQLiteDatabaseQueryBuilder) buildDeletePipeIRsQuery(pipelineIRID int64) (string, []any) {
	return `DELETE FROM pipe_ir WHERE pipeline_ir_id = ?`, []any{pipelineIRID}
}

// buildDeletePipelineIRQuery builds the SQL to delete pipeline_ir by pipeline_id
func (b *SQLiteDatabaseQueryBuilder) buildDeletePipelineIRQuery(pipelineID int64) (string, []any) {
	return `DELETE FROM pipeline_ir WHERE pipeline_id = ?`, []any{pipelineID}
}

// buildDeletePipelineQuery builds the SQL to delete a pipeline by namespace and identifier
func (b *SQLiteDatabaseQueryBuilder) buildDeletePipelineQuery(namespace, identifier string) (string, []any) {
	return `DELETE FROM pipeline WHERE namespace = ? AND identifier = ?`, []any{namespace, identifier}
}

// buildGetPipelineByIDQuery builds the SQL to get a pipeline by id
func (b *SQLiteDatabaseQueryBuilder) buildGetPipelineByIDQuery(pipelineID int64) (string, []any) {
	return `SELECT namespace, identifier FROM pipeline WHERE id = ?`, []any{pipelineID}
}

// buildDistinctNamespacesQuery returns the SQL to get distinct namespaces
func (b *SQLiteDatabaseQueryBuilder) buildDistinctNamespacesQuery() string {
	return "SELECT DISTINCT namespace FROM pipeline"
}

// buildGetPipelinesByNamespaceQuery builds the SQL to get pipelines by namespace
func (b *SQLiteDatabaseQueryBuilder) buildGetPipelinesByNamespaceQuery(namespace string) (string, []any) {
	return `SELECT id, namespace, identifier FROM pipeline WHERE namespace = ? ORDER BY identifier`, []any{namespace}
}

// buildCheckPipelineExistsQuery builds the SQL to check if a pipeline exists
func (b *SQLiteDatabaseQueryBuilder) buildCheckPipelineExistsQuery(namespace, identifier string) (string, []any) {
	return `SELECT id FROM pipeline WHERE namespace = ? AND identifier = ?`, []any{namespace, identifier}
}
