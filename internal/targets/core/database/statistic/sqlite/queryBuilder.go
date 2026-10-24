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

// getCreateTablesSQL returns the SQL to create all statistics tables
func (b *SQLiteDatabaseQueryBuilder) getCreateTablesSQL() string {
	return `
	CREATE TABLE IF NOT EXISTS statistic (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		namespace TEXT NOT NULL,
		pipeline TEXT NOT NULL,
		elapsed_ns INTEGER NOT NULL DEFAULT 0,
		num_of_functions INTEGER NOT NULL DEFAULT 0,
		num_of_channels INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS statistics_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		statistic_id INTEGER NOT NULL,
		FOREIGN KEY (statistic_id) REFERENCES statistic(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS function_statistic (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		statistics_data_id INTEGER NOT NULL,
		"index" INTEGER NOT NULL,
		active INTEGER NOT NULL DEFAULT 0,
		provisions INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (statistics_data_id) REFERENCES statistics_data(id) ON DELETE CASCADE,
		UNIQUE(statistics_data_id, "index")
	);

	CREATE TABLE IF NOT EXISTS pipe_statistic (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		statistics_data_id INTEGER NOT NULL,
		"index" INTEGER NOT NULL,
		pushed INTEGER NOT NULL DEFAULT 0,
		pulled INTEGER NOT NULL DEFAULT 0,
		dropped INTEGER NOT NULL DEFAULT 0,
		breaches INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (statistics_data_id) REFERENCES statistics_data(id) ON DELETE CASCADE,
		UNIQUE(statistics_data_id, "index")
	);

	CREATE TABLE IF NOT EXISTS timing_statistic (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pipe_statistic_id INTEGER NOT NULL,
		min_time_before_pop_ns INTEGER NOT NULL DEFAULT 0,
		max_time_before_pop_ns INTEGER NOT NULL DEFAULT 0,
		average_time_ns INTEGER NOT NULL DEFAULT 0,
		median_time_ns INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (pipe_statistic_id) REFERENCES pipe_statistic(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_statistic_namespace ON statistic(namespace);
	CREATE INDEX IF NOT EXISTS idx_statistic_pipeline ON statistic(pipeline);
	CREATE INDEX IF NOT EXISTS idx_statistic_timestamp ON statistic(timestamp);
	CREATE INDEX IF NOT EXISTS idx_statistic_namespace_pipeline ON statistic(namespace, pipeline);
	CREATE INDEX IF NOT EXISTS idx_statistics_data_statistic_id ON statistics_data(statistic_id);
	CREATE INDEX IF NOT EXISTS idx_function_statistic_data_id ON function_statistic(statistics_data_id);
	CREATE INDEX IF NOT EXISTS idx_pipe_statistic_data_id ON pipe_statistic(statistics_data_id);
	CREATE INDEX IF NOT EXISTS idx_timing_statistic_pipe_id ON timing_statistic(pipe_statistic_id);
	`
}

// buildGetStatisticsQuery builds the SQL to get statistic ids by filter
func (b *SQLiteDatabaseQueryBuilder) buildGetStatisticsQuery(filter string, namespace, pipeline string) (string, []any) {
	b.conditions = b.conditions[:0]
	b.args = b.args[:0]

	if namespace != "" {
		b.conditions = append(b.conditions, "namespace = ?")
		b.args = append(b.args, namespace)
		if pipeline != "" {
			b.conditions = append(b.conditions, "pipeline = ?")
			b.args = append(b.args, pipeline)
		}
	}

	if len(b.conditions) > 0 {
		query := `SELECT id FROM statistic WHERE ` + b.conditions[0]
		for i := 1; i < len(b.conditions); i++ {
			query += " AND " + b.conditions[i]
		}
		return query, b.args
	}

	return `SELECT id FROM statistic`, b.args
}

// buildGetStatisticsDataQuery builds the SQL to get statistics_data by statistic_id
func (b *SQLiteDatabaseQueryBuilder) buildGetStatisticsDataQuery(statisticID int64) (string, []any) {
	return `SELECT id FROM statistics_data WHERE statistic_id = ?`, []any{statisticID}
}

// buildGetFunctionStatisticsQuery builds the SQL to get function_statistics by statistics_data_id
func (b *SQLiteDatabaseQueryBuilder) buildGetFunctionStatisticsQuery(statisticsDataID int64) (string, []any) {
	return `SELECT id, "index", active, provisions FROM function_statistic WHERE statistics_data_id = ? ORDER BY "index"`, []any{statisticsDataID}
}

// buildGetPipeStatisticsQuery builds the SQL to get pipe_statistics by statistics_data_id
func (b *SQLiteDatabaseQueryBuilder) buildGetPipeStatisticsQuery(statisticsDataID int64) (string, []any) {
	return `SELECT id, "index", pushed, pulled, dropped, breaches FROM pipe_statistic WHERE statistics_data_id = ? ORDER BY "index"`, []any{statisticsDataID}
}

// buildGetTimingStatisticsQuery builds the SQL to get timing_statistics by pipe_statistic_id
func (b *SQLiteDatabaseQueryBuilder) buildGetTimingStatisticsQuery(pipeStatisticID int64) (string, []any) {
	return `SELECT id, min_time_before_pop_ns, max_time_before_pop_ns, average_time_ns, median_time_ns FROM timing_statistic WHERE pipe_statistic_id = ?`, []any{pipeStatisticID}
}

// buildInsertStatisticQuery builds the SQL to insert a statistic
func (b *SQLiteDatabaseQueryBuilder) buildInsertStatisticQuery(timestamp, namespace, pipeline string, elapsedNs, numOfFunctions, numOfChannels int64) (string, []any) {
	return `INSERT INTO statistic (timestamp, namespace, pipeline, elapsed_ns, num_of_functions, num_of_channels) VALUES (?, ?, ?, ?, ?, ?)`,
		[]any{timestamp, namespace, pipeline, elapsedNs, numOfFunctions, numOfChannels}
}

// buildInsertStatisticsDataQuery builds the SQL to insert statistics_data
func (b *SQLiteDatabaseQueryBuilder) buildInsertStatisticsDataQuery(statisticID int64) (string, []any) {
	return `INSERT INTO statistics_data (statistic_id) VALUES (?)`, []any{statisticID}
}

// buildInsertFunctionStatisticQuery builds the SQL to insert a function_statistic
func (b *SQLiteDatabaseQueryBuilder) buildInsertFunctionStatisticQuery(statisticsDataID int64, index int, active, provisions int64) (string, []any) {
	return `INSERT INTO function_statistic (statistics_data_id, "index", active, provisions) VALUES (?, ?, ?, ?)`,
		[]any{statisticsDataID, index, active, provisions}
}

// buildInsertPipeStatisticQuery builds the SQL to insert a pipe_statistic
func (b *SQLiteDatabaseQueryBuilder) buildInsertPipeStatisticQuery(statisticsDataID int64, index int, pushed, pulled, dropped, breaches int64) (string, []any) {
	return `INSERT INTO pipe_statistic (statistics_data_id, "index", pushed, pulled, dropped, breaches) VALUES (?, ?, ?, ?, ?, ?)`,
		[]any{statisticsDataID, index, pushed, pulled, dropped, breaches}
}

// buildInsertTimingStatisticQuery builds the SQL to insert a timing_statistic
func (b *SQLiteDatabaseQueryBuilder) buildInsertTimingStatisticQuery(pipeStatisticID int64, minTime, maxTime, avgTime, medianTime int64) (string, []any) {
	return `INSERT INTO timing_statistic (pipe_statistic_id, min_time_before_pop_ns, max_time_before_pop_ns, average_time_ns, median_time_ns) VALUES (?, ?, ?, ?, ?)`,
		[]any{pipeStatisticID, minTime, maxTime, avgTime, medianTime}
}

// buildDeleteStatisticsQuery builds the SQL to delete statistics based on filter
func (b *SQLiteDatabaseQueryBuilder) buildDeleteStatisticsQuery(namespace, pipeline string) (string, []any) {
	b.conditions = b.conditions[:0]
	b.args = b.args[:0]

	if namespace != "" {
		b.conditions = append(b.conditions, "namespace = ?")
		b.args = append(b.args, namespace)
		if pipeline != "" {
			b.conditions = append(b.conditions, "pipeline = ?")
			b.args = append(b.args, pipeline)
		}
	}

	if len(b.conditions) > 0 {
		deleteSQL := `DELETE FROM statistic WHERE ` + b.conditions[0]
		for i := 1; i < len(b.conditions); i++ {
			deleteSQL += " AND " + b.conditions[i]
		}
		return deleteSQL, b.args
	}

	return `DELETE FROM statistic`, b.args
}

// buildDistinctNamespacesQuery returns the SQL to get distinct namespaces
func (b *SQLiteDatabaseQueryBuilder) buildDistinctNamespacesQuery() string {
	return "SELECT DISTINCT namespace FROM statistic"
}

// buildDistinctPipelinesQuery returns the SQL to get distinct pipelines for a namespace
func (b *SQLiteDatabaseQueryBuilder) buildDistinctPipelinesQuery(namespace string) (string, []any) {
	return "SELECT DISTINCT pipeline FROM statistic WHERE namespace = ?", []any{namespace}
}

// buildGetStatisticByIDQuery builds the SQL to get a statistic by id
func (b *SQLiteDatabaseQueryBuilder) buildGetStatisticByIDQuery(statisticID int64) (string, []any) {
	return `SELECT timestamp, namespace, pipeline, elapsed_ns, num_of_functions, num_of_channels FROM statistic WHERE id = ?`, []any{statisticID}
}
