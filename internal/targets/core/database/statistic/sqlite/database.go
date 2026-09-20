// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved.
//
// Source file:  database.go
package sqlite

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/drivers/sqlite"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic"
	"github.com/GabeCordo/ScalingFunctions"
)

////////////////////////////////////////////////////////////////////////
//					SQLite Constants
////////////////////////////////////////////////////////////////////////

const DatabaseName string = "DistributedFunctions"
const CollectionName string = "statistics"

////////////////////////////////////////////////////////////////////////
//					SQLiteDatabase
////////////////////////////////////////////////////////////////////////
//
//		SQLiteDatabase
//			∟ SQLiteDatabaseQueryBuilder	:	owns
//
////////////////////////////////////////////////////////////////////////

// SQLiteDatabase is an implementation of the `Database` interface for the `Statistics`
// struct stored inside SQLite databases.
type SQLiteDatabase struct {
	driver       *sqlite.Driver              // A third party driver for talking to the sqlite db.
	mutex        sync.RWMutex                // Synchronizes access to the sqlite driver.
	path         string                      // The file location of the sqlite database.
	queryBuilder *SQLiteDatabaseQueryBuilder // Holds query building functions.
}

// NewSQLiteDatabase allocates memory for a new `SQLiteDatabase`.
// Use Load() to initialize the database connection.
func NewSQLiteDatabase(path string) (*SQLiteDatabase, error) {
	db := &SQLiteDatabase{
		path:         path,
		queryBuilder: newSQLiteDatabaseQueryBuilder(),
	}
	return db, nil
}

// ensureTable creates the statistics tables if they don't exist.
func (db *SQLiteDatabase) ensureTable() error {
	createTablesSQL := db.queryBuilder.getCreateTablesSQL()
	return db.driver.Exec(createTablesSQL)
}

// IsConnected checks if there is an active connection to the SQLite database.
func (db *SQLiteDatabase) IsConnected() bool {
	if db.driver == nil {
		return false
	}
	return db.driver.IsConnected()
}

// Get retrieves zero-to-many `Statistics` records from the SQLite database.
func (db *SQLiteDatabase) Get(filter database.Filter) (records []*ScalingFunctions.Statistics) {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		return records
	}

	query, args := db.queryBuilder.buildGetStatisticsQuery("", filter.Namespace, filter.Pipeline)
	rows, err := db.driver.Query(query, args...)
	if err != nil {
		log.Printf("Error querying statistics: %v", err)
		return records
	}
	defer rows.Close()

	var id int64

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Printf("Error scanning statistic row: %v", err)
			continue
		}

		fullStats, err := db.getStatisticsByID(id)
		if err != nil {
			log.Printf("Error getting full statistics for id %d: %v", id, err)
			continue
		}

		if fullStats != nil {
			records = append(records, fullStats)
		}
	}

	return records
}

// getStatisticsByID retrieves a Statistics record by its database ID
func (db *SQLiteDatabase) getStatisticsByID(statisticId int64) (*ScalingFunctions.Statistics, error) {

	query, args := db.queryBuilder.buildGetStatisticByIDQuery(statisticId)
	row := db.driver.QueryRow(query, args...)

	var timestampStr, namespace, pipeline string
	var elapsedNs, numOfFunctions, numOfChannels int64

	err := row.Scan(&timestampStr, &namespace, &pipeline, &elapsedNs, &numOfFunctions, &numOfChannels)
	if err != nil {
		if sqlite.IsNoRows(err) {
			return nil, fmt.Errorf("statistic with id %d not found", statisticId)
		}
		return nil, fmt.Errorf("failed to scan statistic: %w", err)
	}

	query, args = db.queryBuilder.buildGetStatisticsDataQuery(statisticId)
	dataRow := db.driver.QueryRow(query, args...)

	var statisticsDataID int64
	err = dataRow.Scan(&statisticsDataID)
	if err != nil {
		if sqlite.IsNoRows(err) {
			// No statistics_data means no nested data
			return &ScalingFunctions.Statistics{}, nil
		}
		return nil, fmt.Errorf("failed to scan statistics_data: %w", err)
	}

	query, args = db.queryBuilder.buildGetFunctionStatisticsQuery(statisticsDataID)
	functionRows, err := db.driver.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query function_statistics: %w", err)
	}
	defer functionRows.Close()

	functions := make([]ScalingFunctions.FunctionStatistic, numOfFunctions)

	var id, index int64
	var active, provisions int64

	for functionRows.Next() {
		err := functionRows.Scan(&id, &index, &active, &provisions)
		if err != nil {
			log.Printf("Error scanning function_statistic row: %v", err)
			continue
		}

		if index >= 0 && index < int64(len(functions)) {
			functions[index] = ScalingFunctions.FunctionStatistic{
				Active:     uint32(active),
				Provisions: uint64(provisions),
			}
		}
	}

	query, args = db.queryBuilder.buildGetPipeStatisticsQuery(statisticsDataID)
	pipeRows, err := db.driver.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pipe_statistics: %w", err)
	}
	defer pipeRows.Close()

	pipes := make([]ScalingFunctions.PipeStatistic, numOfChannels)
	for pipeRows.Next() {
		var id, pipeStatisticID, index int64
		var pushed, pulled, dropped, breaches int64

		err := pipeRows.Scan(&id, &index, &pushed, &pulled, &dropped, &breaches)
		if err != nil {
			log.Printf("Error scanning pipe_statistic row: %v", err)
			continue
		}

		if index >= 0 && index < int64(len(pipes)) {
			pipes[index] = ScalingFunctions.PipeStatistic{
				Pushed:   uint64(pushed),
				Pulled:   uint64(pulled),
				Dropped:  uint64(dropped),
				Breaches: uint64(breaches),
			}

			// Get timing statistics for this pipe
			query, args = db.queryBuilder.buildGetTimingStatisticsQuery(id)
			timingRow := db.driver.QueryRow(query, args...)

			var minTime, maxTime, avgTime, medianTime int64
			err := timingRow.Scan(&pipeStatisticID, &minTime, &maxTime, &avgTime, &medianTime)
			if err != nil {
				if !sqlite.IsNoRows(err) {
					log.Printf("Error scanning timing_statistic row: %v", err)
				}
			} else {
				pipes[index].Timing = ScalingFunctions.TimingStatistics{
					MinTimeBeforePop: time.Duration(minTime),
					MaxTimeBeforePop: time.Duration(maxTime),
					AverageTime:      time.Duration(avgTime),
					MedianTime:       time.Duration(medianTime),
				}
			}
		}
	}

	return &ScalingFunctions.Statistics{
		NumOfFunctions: uint16(numOfFunctions),
		Functions:      functions,
		NumOfChannels:  uint16(numOfChannels),
		Pipes:          pipes,
	}, nil
}

// Create inserts a `Statistics` record to the SQLite database.
func (db *SQLiteDatabase) Create(filter database.Filter, record *ScalingFunctions.Statistics) (*ScalingFunctions.Statistics, error) {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return nil, database.NotConnected
	}

	// Insert the statistic
	query, args := db.queryBuilder.buildInsertStatisticQuery(
		time.Now().Format(time.RFC3339),
		filter.Namespace,
		filter.Pipeline,
		0, // elapsed_ns
		int64(record.NumOfFunctions),
		int64(record.NumOfChannels),
	)
	result, err := db.driver.ExecResult(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert statistic: %w", err)
	}

	statisticID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Insert statistics_data
	query, args = db.queryBuilder.buildInsertStatisticsDataQuery(statisticID)
	dataResult, err := db.driver.ExecResult(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert statistics_data: %w", err)
	}

	statisticsDataID, err := dataResult.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics_data id: %w", err)
	}

	// Insert function statistics
	for idx, function := range record.Functions {
		query, args = db.queryBuilder.buildInsertFunctionStatisticQuery(
			statisticsDataID,
			idx,
			int64(function.Active),
			int64(function.Provisions),
		)
		err := db.driver.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert function_statistic: %w", err)
		}
	}

	// Insert pipe statistics
	for idx, pipe := range record.Pipes {
		// Insert pipe_statistic
		query, args = db.queryBuilder.buildInsertPipeStatisticQuery(
			statisticsDataID,
			idx,
			int64(pipe.Pushed),
			int64(pipe.Pulled),
			int64(pipe.Dropped),
			int64(pipe.Breaches),
		)
		pipeResult, err := db.driver.ExecResult(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert pipe_statistic: %w", err)
		}

		pipeStatisticID, err := pipeResult.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get pipe_statistic id: %w", err)
		}

		// Insert timing_statistic
		query, args = db.queryBuilder.buildInsertTimingStatisticQuery(
			pipeStatisticID,
			int64(pipe.Timing.MinTimeBeforePop),
			int64(pipe.Timing.MaxTimeBeforePop),
			int64(pipe.Timing.AverageTime),
			int64(pipe.Timing.MedianTime),
		)
		err = db.driver.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert timing_statistic: %w", err)
		}
	}

	return record, nil
}

// Replace is not implemented for the statistic.SQLiteDatabase
func (db *SQLiteDatabase) Replace(filter database.Filter, record *ScalingFunctions.Statistics) error {
	return database.NotImplemented
}

// Delete removes Statistics records from the SQLite database
func (db *SQLiteDatabase) Delete(filter database.Filter) error {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return database.NotConnected
	}

	// Delete by namespace and pipeline - cascade will handle the rest
	query, args := db.queryBuilder.buildDeleteStatisticsQuery(filter.Namespace, filter.Pipeline)
	return db.driver.Exec(query, args...)
}

// Distinct returns distinct values for a given field
func (db *SQLiteDatabase) Distinct(filter database.Filter) ([]any, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		return nil, database.NotConnected
	}

	var results []any

	// If namespace is specified, return distinct pipelines
	if filter.Namespace != "" {
		query, args := db.queryBuilder.buildDistinctPipelinesQuery(filter.Namespace)
		rows, err := db.driver.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to query distinct pipelines: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var pipeline string
			err := rows.Scan(&pipeline)
			if err != nil {
				log.Printf("Error scanning distinct pipeline: %v", err)
				continue
			}
			results = append(results, pipeline)
		}
	} else {
		// Return distinct namespaces
		query := db.queryBuilder.buildDistinctNamespacesQuery()
		rows, err := db.driver.Query(query)
		if err != nil {
			return nil, fmt.Errorf("failed to query distinct namespaces: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var namespace string
			err := rows.Scan(&namespace)
			if err != nil {
				log.Printf("Error scanning distinct namespace: %v", err)
				continue
			}
			results = append(results, namespace)
		}
	}

	return results, nil
}

// Save saves the database to disk (for SQLite, this is automatic, but we can implement backup)
func (db *SQLiteDatabase) Save(path string) error {

	// The Save() function shall not be called while there is an ongoing write operation
	// Insert(), Replace(), or Delete()
	//
	// The Save() function shall run in parallel with lookup operations such as Get()
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		return database.NotConnected
	}

	// All modifications persist so no explicit save needs to be performed.
	return nil
}

// Load loads data from disk by initializing the SQLite driver and ensuring tables exist
func (db *SQLiteDatabase) Load(path string) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	// `path` shall not be an empty string since we cannot load a sqlite database from an empty string.
	if db.path == "" {
		return errors.New("no database path specified")
	}

	dir := filepath.Dir(db.path)

	// `path` shall not be a non-absolute path.
	// 		1) "" implies `path` was a file name without a path.
	//		2) "." implies `path` was a file name with a relative path.
	if dir == "" || dir == "." {
		return fmt.Errorf("path '%s' must be an absolute path", path)
	}

	// if `dir` does not exist then create the directory.
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// if `db.path` does not exist then create the sqlite file.
	driver, err := sqlite.NewDriver(db.path)

	if err != nil {
		return fmt.Errorf("failed to create SQLite driver: %w", err)
	}

	if driver == nil {
		return fmt.Errorf("creation of the SQLite driver returned a nullptr")
	}

	db.driver = driver

	err = db.ensureTable()
	if err != nil {
		return fmt.Errorf("failed to ensure statistics table exists: %w", err)
	}

	return nil
}

// Print outputs the Statistics records from the database to the console
func (db *SQLiteDatabase) Print() {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		fmt.Println("Database is not connected")
		return
	}

	// Get all namespaces
	namespaces, err := db.Distinct(database.Filter{})
	if err != nil {
		fmt.Printf("Error getting namespaces: %v\n", err)
		return
	}

	for _, ns := range namespaces {
		namespace := ns.(string)
		fmt.Printf("├─ %s\n", namespace)

		// Get all pipelines in this namespace
		pipelines, err := db.Distinct(database.Filter{Namespace: namespace})
		if err != nil {
			fmt.Printf("  Error getting pipelines for %s: %v\n", namespace, err)
			continue
		}

		for _, pl := range pipelines {
			pipeline := pl.(string)
			// Get statistics for this namespace and pipeline
			stats := db.Get(database.Filter{Namespace: namespace, Pipeline: pipeline})
			fmt.Printf("|   ├─ %s (num of records: %d)\n", pipeline, len(stats))
		}
	}
}

// Close closes the database connection
func (db *SQLiteDatabase) Close() error {
	if db.driver == nil {
		return nil
	}
	return db.driver.Close()
}

// CreateWithStatistic creates a statistic record with full metadata from the statistic.Statistic type
func (db *SQLiteDatabase) CreateWithStatistic(filter database.Filter, record *statistic.Statistic) (*ScalingFunctions.Statistics, error) {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return nil, database.NotConnected
	}

	// Insert the statistic
	query, args := db.queryBuilder.buildInsertStatisticQuery(
		record.Timestamp.Format(time.RFC3339),
		record.Namespace,
		record.Pipeline,
		int64(record.Elapsed),
		int64(record.Data.NumOfFunctions),
		int64(record.Data.NumOfChannels),
	)
	result, err := db.driver.ExecResult(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert statistic: %w", err)
	}

	statisticID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Insert statistics_data
	query, args = db.queryBuilder.buildInsertStatisticsDataQuery(statisticID)
	dataResult, err := db.driver.ExecResult(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert statistics_data: %w", err)
	}

	statisticsDataID, err := dataResult.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics_data id: %w", err)
	}

	// Insert function statistics
	for idx, function := range record.Data.Functions {
		query, args = db.queryBuilder.buildInsertFunctionStatisticQuery(
			statisticsDataID,
			idx,
			int64(function.Active),
			int64(function.Provisions),
		)
		err := db.driver.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert function_statistic: %w", err)
		}
	}

	// Insert pipe statistics
	for idx, pipe := range record.Data.Pipes {
		// Insert pipe_statistic
		query, args = db.queryBuilder.buildInsertPipeStatisticQuery(
			statisticsDataID,
			idx,
			int64(pipe.Pushed),
			int64(pipe.Pulled),
			int64(pipe.Dropped),
			int64(pipe.Breaches),
		)
		pipeResult, err := db.driver.ExecResult(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert pipe_statistic: %w", err)
		}

		pipeStatisticID, err := pipeResult.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get pipe_statistic id: %w", err)
		}

		// Insert timing_statistic
		query, args = db.queryBuilder.buildInsertTimingStatisticQuery(
			pipeStatisticID,
			int64(pipe.Timing.MinTimeBeforePop),
			int64(pipe.Timing.MaxTimeBeforePop),
			int64(pipe.Timing.AverageTime),
			int64(pipe.Timing.MedianTime),
		)
		err = db.driver.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert timing_statistic: %w", err)
		}
	}

	return record.Data, nil
}
