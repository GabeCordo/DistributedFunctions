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

	"github.com/GabeCordo/DistributedFunctions/internal/shared/drivers/sqlite"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

////////////////////////////////////////////////////////////////////////
//						   SQLite Constants
////////////////////////////////////////////////////////////////////////

const DatabaseName string = "DistributedFunctions"
const CollectionName string = "jobs"

////////////////////////////////////////////////////////////////////////
//							SQLightDatabase
////////////////////////////////////////////////////////////////////////
//
//		SQLiteDatabase
//			∟ SQLiteDatabaseQueryBuilder	:	owns
//
////////////////////////////////////////////////////////////////////////

// SQLiteDatabase is an implementation of the `Database` interface for the `Job`
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

// ensureTable creates the jobs, job_intervals, and job_metadata tables if they don't exist.
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

// Get retrieves zero-to-many `Job` records from the SQLite database.
func (db *SQLiteDatabase) Get(filter database.Filter) (jobs []*job.Job) {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		return jobs
	}

	// the `buildGetQueryWithArgs` function uses two left joins creating a cartesian product
	// example. a job with 2 interval rows and 3 metadata rows will have 6 rows.
	query, args := db.queryBuilder.buildGetQueryWithArgs(filter)
	rows, err := db.driver.Query(query, args...)
	if err != nil {
		log.Printf("Error querying jobs: %v", err)
		return jobs
	}
	defer rows.Close()

	// the `jobMap` is used to group rows with the same `jobId`
	jobMap := make(map[int64]*job.Job)

	var jobID int64
	var identifier, namespace, pipeline string
	var minute, hour, day, month int
	var key, value sqlite.NullString

	for rows.Next() {

		err := rows.Scan(&jobID, &identifier, &namespace, &pipeline, &minute, &hour, &day, &month, &key, &value)
		if err != nil {
			log.Printf("Error scanning job row: %v", err)
			continue
		}

		j, exists := jobMap[jobID]
		if !exists {
			j = &job.Job{
				Identifier: identifier,
				Namespace:  namespace,
				Pipeline:   pipeline,
				Metadata:   make(map[string]string),
			}
			jobMap[jobID] = j
		}

		j.Interval = database.Interval{
			Minute: minute,
			Hour:   hour,
			Day:    day,
			Month:  month,
		}

		if key.Valid() && value.Valid() {
			j.Metadata[key.String()] = value.String()
		}
	}

	for _, j := range jobMap {
		jobs = append(jobs, j)
	}

	return jobs
}

// Create inserts a `Job` record to the SQLite database.
func (db *SQLiteDatabase) Create(filter database.Filter, record *job.Job) (string, error) {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return "", database.NotConnected
	}

	// Check if the job already exists
	query, args := db.queryBuilder.buildGetQueryWithArgs(filter)
	rows, err := db.driver.Query(query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to check for existing job: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return "", database.AlreadyExists
	}

	// Since there are three table insertions tied to the creation of a `Job`
	// record: (1) JobTable, (2) JobIntervalTable, (3) JobMetadataTable we need
	// to group the insertions into an atomic transaction.
	//
	// The transaction will rollback with `tx.Rollback()` if any of the three
	// insertions fail.
	tx, err := db.driver.BeginTransaction()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query, args = db.queryBuilder.buildInsertJobQuery(record.Identifier, record.Namespace, record.Pipeline)
	result, err := tx.Exec(query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to insert job: %w", err)
	}

	jobId, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get last insert id which represents the job id: %w", err)
	}

	query, args = db.queryBuilder.buildInsertIntervalQuery(jobId, record.Interval)
	_, err = tx.Exec(query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to insert job interval: %w", err)
	}

	for key, value := range record.Metadata {
		query, args = db.queryBuilder.buildInsertMetadataQuery(jobId, key, value)
		_, err = tx.Exec(query, args...)
		if err != nil {
			return "", fmt.Errorf("failed to insert job metadata: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return record.Identifier, nil
}

// Replace replaces an existing `Job` record in the SQLite database.
func (db *SQLiteDatabase) Replace(filter database.Filter, record *job.Job) error {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return database.NotConnected
	}

	query, args := db.queryBuilder.buildGetJobIDQuery(filter)
	if query == "" {
		return errors.New("filter must specify at least one of: identifier, namespace, pipeline, or interval for replace")
	}

	var jobID int64
	err := db.driver.QueryRow(query, args...).Scan(&jobID)
	if err != nil {
		if sqlite.IsNoRows(err) {
			return errors.New("no job found matching the filter for replace")
		}
		return fmt.Errorf("failed to find job for replace: %w", err)
	}

	// Since there are three table updates tied to the replacement of a `Job`
	// record: (1) JobTable, (2) JobIntervalTable, (3) JobMetadataTable we need
	// to group the updates into an atomic transaction.
	//
	// The transaction will rollback with `tx.Rollback()` if any of the three
	// updates fail.
	tx, err := db.driver.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query, args = db.queryBuilder.buildUpdateJobQuery(record, jobID)
	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	query, args = db.queryBuilder.buildUpdateIntervalQuery(record.Interval, jobID)
	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update job interval: %w", err)
	}

	query, args = db.queryBuilder.buildDeleteMetadataQuery(jobID)
	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete job metadata: %w", err)
	}

	for key, value := range record.Metadata {
		query, args = db.queryBuilder.buildInsertMetadataQuery(jobID, key, value)
		_, err = tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("failed to insert job metadata: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete removes Job records from the SQLite database
func (db *SQLiteDatabase) Delete(filter database.Filter) error {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return database.NotConnected
	}

	query, args := db.queryBuilder.buildDeleteJobsQuery(filter)
	return db.driver.Exec(query, args...)
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

	// `path` shall not be an empty string since we cannot load a sqlite databased from an empty string.
	if db.path == "" {
		return errors.New("no database path specified")
	}

	dir := filepath.Dir(db.path)

	// `path` shall not be a non-absolute path.
	// 		1) "" implies `path` was a file name without a path.
	//		2) '.' implies `path` was a file name with a relative path.
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
		return fmt.Errorf("failed to ensure jobs table exists: %w", err)
	}

	return nil
}

// Print outputs the Job records from the database to the console
func (db *SQLiteDatabase) Print() {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		fmt.Println("Database is not connected")
		return
	}

	jobs := db.Get(database.Filter{})
	for _, jobItem := range jobs {
		fmt.Printf("├─ %s\n", jobItem.ToString())
	}
}

// Close closes the database connection
func (db *SQLiteDatabase) Close() error {
	if db.driver == nil {
		return nil
	}
	return db.driver.Close()
}
