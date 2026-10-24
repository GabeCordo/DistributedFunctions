// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved.
//
// Source file:  queryBuilder.go
package sqlite

import (
	"fmt"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

////////////////////////////////////////////////////////////////////////
//						SQLightDatabaseQueryBuilder
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
		args:       make([]any, 0, 7),
	}
}

// getCreateTablesSQL returns the SQL to create all tables
func (b *SQLiteDatabaseQueryBuilder) getCreateTablesSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		identifier TEXT NOT NULL UNIQUE,
		namespace TEXT NOT NULL,
		pipeline TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS job_intervals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		minute INTEGER NOT NULL DEFAULT 0,
		hour INTEGER NOT NULL DEFAULT 0,
		day INTEGER NOT NULL DEFAULT 0,
		month INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS job_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
		UNIQUE(job_id, key)
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_namespace ON jobs(namespace);
	CREATE INDEX IF NOT EXISTS idx_jobs_pipeline ON jobs(pipeline);
	CREATE INDEX IF NOT EXISTS idx_jobs_identifier ON jobs(identifier);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_jobs_unique ON jobs(namespace, pipeline);
	CREATE INDEX IF NOT EXISTS idx_job_intervals_job_id ON job_intervals(job_id);
	CREATE INDEX IF NOT EXISTS idx_job_metadata_job_id ON job_metadata(job_id);
	`
}

// buildGetQueryWithArgs builds the SQL query and arguments based on the filter
func (b *SQLiteDatabaseQueryBuilder) buildGetQueryWithArgs(filter database.Filter) (string, []any) {
	const selectJobsQuery = `SELECT 
		j.id, j.identifier, j.namespace, j.pipeline,
		ji.minute, ji.hour, ji.day, ji.month,
		jm.key, jm.value
	FROM jobs j
	LEFT JOIN job_intervals ji ON j.id = ji.job_id
	LEFT JOIN job_metadata jm ON j.id = jm.job_id`

	// Reset slices to reuse pre-allocated capacity
	b.conditions = b.conditions[:0]
	b.args = b.args[:0]

	if filter.UseIdentifier() {
		b.conditions = append(b.conditions, "j.identifier = ?")
		b.args = append(b.args, filter.Identifier)
	}
	if filter.UseNamespace() {
		b.conditions = append(b.conditions, "j.namespace = ?")
		b.args = append(b.args, filter.Namespace)
	}
	if filter.UsePipeline() {
		b.conditions = append(b.conditions, "j.pipeline = ?")
		b.args = append(b.args, filter.Pipeline)
	}
	if filter.UseInterval() {
		b.conditions = append(b.conditions, "ji.minute = ? AND ji.hour = ? AND ji.day = ? AND ji.month = ?")
		b.args = append(b.args, filter.Interval.Minute, filter.Interval.Hour, filter.Interval.Day, filter.Interval.Month)
	}

	if len(b.conditions) > 0 {
		query := fmt.Sprintf("%s WHERE %s", selectJobsQuery, b.conditions[0])
		for i := 1; i < len(b.conditions); i++ {
			query += " AND " + b.conditions[i]
		}
		return query, b.args
	}

	return selectJobsQuery, b.args
}

// buildInsertJobQuery builds the SQL to insert a job
func (b *SQLiteDatabaseQueryBuilder) buildInsertJobQuery(identifier, namespace, pipeline string) (string, []any) {
	return `INSERT INTO jobs (identifier, namespace, pipeline) VALUES (?, ?, ?)`, []any{identifier, namespace, pipeline}
}

// buildInsertIntervalQuery builds the SQL to insert a job interval
func (b *SQLiteDatabaseQueryBuilder) buildInsertIntervalQuery(jobID int64, interval database.Interval) (string, []any) {
	return `INSERT INTO job_intervals (job_id, minute, hour, day, month) VALUES (?, ?, ?, ?, ?)`,
		[]any{jobID, interval.Minute, interval.Hour, interval.Day, interval.Month}
}

// buildInsertMetadataQuery builds the SQL to insert a job metadata entry
func (b *SQLiteDatabaseQueryBuilder) buildInsertMetadataQuery(jobID int64, key, value string) (string, []any) {
	return `INSERT INTO job_metadata (job_id, key, value) VALUES (?, ?, ?)`, []any{jobID, key, value}
}

// buildGetJobIDQuery builds the SQL to find a job ID by filter
func (b *SQLiteDatabaseQueryBuilder) buildGetJobIDQuery(filter database.Filter) (string, []any) {
	b.conditions = b.conditions[:0]
	b.args = b.args[:0]

	if filter.UseIdentifier() {
		b.conditions = append(b.conditions, "identifier = ?")
		b.args = append(b.args, filter.Identifier)
	}
	if filter.UseNamespace() {
		b.conditions = append(b.conditions, "namespace = ?")
		b.args = append(b.args, filter.Namespace)
	}
	if filter.UsePipeline() {
		b.conditions = append(b.conditions, "pipeline = ?")
		b.args = append(b.args, filter.Pipeline)
	}

	if len(b.conditions) == 0 {
		return "", nil
	}

	query := fmt.Sprintf("SELECT id FROM jobs WHERE %s", b.conditions[0])
	for i := 1; i < len(b.conditions); i++ {
		query += " AND " + b.conditions[i]
	}
	return query, b.args
}

// buildUpdateJobQuery builds the SQL to update a job
func (b *SQLiteDatabaseQueryBuilder) buildUpdateJobQuery(record *job.Job, jobID int64) (string, []any) {
	return `UPDATE jobs SET identifier = ?, namespace = ?, pipeline = ? WHERE id = ?`,
		[]any{record.Identifier, record.Namespace, record.Pipeline, jobID}
}

// buildUpdateIntervalQuery builds the SQL to update a job interval
func (b *SQLiteDatabaseQueryBuilder) buildUpdateIntervalQuery(interval database.Interval, jobID int64) (string, []any) {
	return `UPDATE job_intervals SET minute = ?, hour = ?, day = ?, month = ? WHERE job_id = ?`,
		[]any{interval.Minute, interval.Hour, interval.Day, interval.Month, jobID}
}

// buildDeleteMetadataQuery builds the SQL to delete metadata for a job
func (b *SQLiteDatabaseQueryBuilder) buildDeleteMetadataQuery(jobID int64) (string, []any) {
	return `DELETE FROM job_metadata WHERE job_id = ?`, []any{jobID}
}

// buildDeleteJobsQuery builds the SQL to delete jobs based on filter
func (b *SQLiteDatabaseQueryBuilder) buildDeleteJobsQuery(filter database.Filter) (string, []any) {
	deleteSQL := "DELETE FROM jobs"
	b.conditions = b.conditions[:0]
	b.args = b.args[:0]

	if filter.UseIdentifier() {
		b.conditions = append(b.conditions, "identifier = ?")
		b.args = append(b.args, filter.Identifier)
	}
	if filter.UseNamespace() {
		b.conditions = append(b.conditions, "namespace = ?")
		b.args = append(b.args, filter.Namespace)
	}
	if filter.UsePipeline() {
		b.conditions = append(b.conditions, "pipeline = ?")
		b.args = append(b.args, filter.Pipeline)
	}
	if filter.UseInterval() {
		deleteSQL = `DELETE FROM jobs WHERE id IN (SELECT job_id FROM job_intervals WHERE minute = ? AND hour = ? AND day = ? AND month = ?)`
		b.args = append(b.args, filter.Interval.Minute, filter.Interval.Hour, filter.Interval.Day, filter.Interval.Month)
		if len(b.conditions) > 0 {
			b.conditions = b.conditions[:0]
			b.args = b.args[:0]
		}
	}

	if len(b.conditions) > 0 {
		deleteSQL += " WHERE " + b.conditions[0]
		for i := 1; i < len(b.conditions); i++ {
			deleteSQL += " AND " + b.conditions[i]
		}
	}

	return deleteSQL, b.args
}
