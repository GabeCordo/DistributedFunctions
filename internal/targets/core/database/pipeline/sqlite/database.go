// Package sqlite
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved.
//
// Source file:  database.go
package sqlite

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/GabeCordo/ScalingFunctions"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/drivers/sqlite"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
)

////////////////////////////////////////////////////////////////////////
//					SQLite Constants
////////////////////////////////////////////////////////////////////////

const DatabaseName string = "DistributedFunctions"
const CollectionName string = "pipelines"

////////////////////////////////////////////////////////////////////////
//					SQLiteDatabase
////////////////////////////////////////////////////////////////////////
//
//		SQLiteDatabase
//			∟ SQLiteDatabaseQueryBuilder	:	owns
//
////////////////////////////////////////////////////////////////////////

// SQLiteDatabase is an implementation of the `Database` interface for the `Pipeline`
// struct stored inside SQLite databases.
type SQLiteDatabase struct {
	driver       *sqlite.Driver                       // A third party driver for talking to the sqlite db.
	mutex        sync.RWMutex                         // Synchronizes access to the sqlite driver.
	path         string                               // The file location of the sqlite database.
	queryBuilder *SQLiteDatabaseQueryBuilder         // Holds query building functions.
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

// ensureTable creates the pipeline tables if they don't exist.
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

// Get retrieves zero-to-many `PipelineIR` records from the SQLite database.
func (db *SQLiteDatabase) Get(filter database.Filter) (records []*ScalingFunctions.PipelineIR) {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		return records
	}

	// Get pipelines by namespace
	if filter.Namespace != "" {
		query, args := db.queryBuilder.buildGetPipelinesByNamespaceQuery(filter.Namespace)
		rows, err := db.driver.Query(query, args...)
		if err != nil {
			log.Printf("Error querying pipelines: %v", err)
			return records
		}
		defer rows.Close()

		for rows.Next() {
			var pipelineID int64
			var namespace, identifier string

			if err := rows.Scan(&pipelineID, &namespace, &identifier); err != nil {
				log.Printf("Error scanning pipeline row: %v", err)
				continue
			}

			// Get the pipeline IR for this pipeline
			pipelineIR, err := db.getPipelineIRByID(pipelineID)
			if err != nil {
				log.Printf("Error getting pipeline IR for pipeline %d: %v", pipelineID, err)
				continue
			}

			if pipelineIR != nil {
				records = append(records, pipelineIR)
			}
		}
	}

	return records
}

// getPipelineIRByID retrieves a PipelineIR by pipeline database ID
func (db *SQLiteDatabase) getPipelineIRByID(pipelineID int64) (*ScalingFunctions.PipelineIR, error) {
	// Get pipeline metadata
	query, args := db.queryBuilder.buildGetPipelineByIDQuery(pipelineID)
	row := db.driver.QueryRow(query, args...)

	var namespace, identifier string
	if err := row.Scan(&namespace, &identifier); err != nil {
		if sqlite.IsNoRows(err) {
			return nil, fmt.Errorf("pipeline with id %d not found", pipelineID)
		}
		return nil, fmt.Errorf("failed to scan pipeline: %w", err)
	}

	// Get pipeline_ir
	query, args = db.queryBuilder.buildGetPipelineIRQuery(pipelineID)
	pipelineIRRow := db.driver.QueryRow(query, args...)

	var pipelineIRID int64
	var onCrashStr, onStartupStr, onTeardownStr sqlite.NullString
	if err := pipelineIRRow.Scan(&pipelineIRID, &pipelineID, &onCrashStr, &onStartupStr, &onTeardownStr); err != nil {
		if sqlite.IsNoRows(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan pipeline_ir: %w", err)
	}

	// Get function_irs
	query, args = db.queryBuilder.buildGetFunctionIRsQuery(pipelineIRID)
	functionRows, err := db.driver.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query function_irs: %w", err)
	}
	defer functionRows.Close()

	var functions []ScalingFunctions.FunctionIR
	for functionRows.Next() {
		var function ScalingFunctions.FunctionIR
		var id, functionPipelineIRID int64
		var module, identifier, fromPipe, toPipe, parameters, returns string
		var startWith, maximum int
		var waitBefore, staticMountInt int64

		if err := functionRows.Scan(&id, &functionPipelineIRID, &module, &identifier, &fromPipe, &toPipe, &startWith, &waitBefore, &maximum, &parameters, &returns, &staticMountInt); err != nil {
			log.Printf("Error scanning function_ir row: %v", err)
			continue
		}

		function.Module = module
		function.Identifier = identifier
		function.From = fromPipe
		function.To = toPipe
		function.StartWith = uint16(startWith)
		function.WaitBefore = waitBefore != 0
		function.Maximum = uint16(maximum)

		// Parse parameters
		if parameters != "" && parameters != "[]" {
			if err := json.Unmarshal([]byte(parameters), &function.Parameters); err != nil {
				log.Printf("Error unmarshaling parameters: %v", err)
			}
		}

		// Parse returns
		if returns != "" && returns != "[]" {
			if err := json.Unmarshal([]byte(returns), &function.Returns); err != nil {
				log.Printf("Error unmarshaling returns: %v", err)
			}
		}

		function.Metadata.StaticMount = staticMountInt != 0

		functions = append(functions, function)
	}

	// Get pipe_irs
	query, args = db.queryBuilder.buildGetPipeIRsQuery(pipelineIRID)
	pipeRows, err := db.driver.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pipe_irs: %w", err)
	}
	defer pipeRows.Close()

	var pipes []ScalingFunctions.PipeIR
	for pipeRows.Next() {
		var pipe ScalingFunctions.PipeIR
		var id, pipePipelineIRID int64
		var identifier string
		var threshold int
		var growthFactor float64

		if err := pipeRows.Scan(&id, &pipePipelineIRID, &identifier, &threshold, &growthFactor); err != nil {
			log.Printf("Error scanning pipe_ir row: %v", err)
			continue
		}

		pipe.Identifier = identifier
		pipe.Threshold = uint32(threshold)
		pipe.GrowthFactor = growthFactor

		pipes = append(pipes, pipe)
	}

	// Build the PipelineIR
	pipelineIR := &ScalingFunctions.PipelineIR{
		Identifier: identifier,
		Functions:  functions,
		Pipes:      pipes,
	}

	if onCrashStr.Valid() {
		pipelineIR.OnCrash = ScalingFunctions.OnCrash(onCrashStr.String())
	}
	if onStartupStr.Valid() {
		pipelineIR.OnStartup = onStartupStr.String()
	}
	if onTeardownStr.Valid() {
		pipelineIR.OnTeardown = onTeardownStr.String()
	}

	return pipelineIR, nil
}

// Create inserts a `PipelineIR` record to the SQLite database.
func (db *SQLiteDatabase) Create(filter database.Filter, record *ScalingFunctions.PipelineIR) (string, error) {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return "", database.NotConnected
	}

	if filter.Namespace == "" {
		return "", errors.New("filter.Namespace is required")
	}

	if filter.Identifier == "" {
		return "", errors.New("filter.Identifier is required")
	}

	// Check if the pipeline already exists
	query, args := db.queryBuilder.buildCheckPipelineExistsQuery(filter.Namespace, filter.Identifier)
	row := db.driver.QueryRow(query, args...)

	var existingID int64
	if err := row.Scan(&existingID); err != nil {
		if !sqlite.IsNoRows(err) {
			return "", fmt.Errorf("failed to check for existing pipeline: %w", err)
		}
	} else {
		return "", database.AlreadyExists
	}

	// Insert the pipeline
	query, args = db.queryBuilder.buildInsertPipelineQuery(filter.Namespace, filter.Identifier)
	result, err := db.driver.ExecResult(query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to insert pipeline: %w", err)
	}

	pipelineID, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Insert the pipeline_ir
	query, args = db.queryBuilder.buildInsertPipelineIRQuery(
		pipelineID,
		string(record.OnCrash),
		record.OnStartup,
		record.OnTeardown,
	)
	result, err = db.driver.ExecResult(query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to insert pipeline_ir: %w", err)
	}

	pipelineIRID, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get pipeline_ir id: %w", err)
	}

	// Insert function_irs
	for _, function := range record.Functions {
		// Marshal parameters and returns
		parametersJSON, err := json.Marshal(function.Parameters)
		if err != nil {
			return "", fmt.Errorf("failed to marshal function parameters: %w", err)
		}
		returnsJSON, err := json.Marshal(function.Returns)
		if err != nil {
			return "", fmt.Errorf("failed to marshal function returns: %w", err)
		}

		query, args = db.queryBuilder.buildInsertFunctionIRQuery(
			pipelineIRID,
			function.Module,
			function.Identifier,
			function.From,
			function.To,
			int(function.StartWith),
			boolToInt(function.WaitBefore),
			int(function.Maximum),
			string(parametersJSON),
			string(returnsJSON),
			function.Metadata.StaticMount,
		)
		if err := db.driver.Exec(query, args...); err != nil {
			return "", fmt.Errorf("failed to insert function_ir: %w", err)
		}
	}

	// Insert pipe_irs
	for _, pipe := range record.Pipes {
		query, args = db.queryBuilder.buildInsertPipeIRQuery(
			pipelineIRID,
			pipe.Identifier,
			int(pipe.Threshold),
			pipe.GrowthFactor,
		)
		if err := db.driver.Exec(query, args...); err != nil {
			return "", fmt.Errorf("failed to insert pipe_ir: %w", err)
		}
	}

	return filter.Identifier, nil
}

// Replace replaces an existing `PipelineIR` record in the SQLite database.
func (db *SQLiteDatabase) Replace(filter database.Filter, record *ScalingFunctions.PipelineIR) error {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return database.NotConnected
	}

	if filter.Namespace == "" {
		return errors.New("filter.Namespace is required")
	}

	if filter.Identifier == "" {
		return errors.New("filter.Identifier is required")
	}

	// Get the pipeline ID
	query, args := db.queryBuilder.buildGetPipelineQuery(filter.Namespace, filter.Identifier)
	row := db.driver.QueryRow(query, args...)

	var pipelineID int64
	if err := row.Scan(&pipelineID, &filter.Namespace, &filter.Identifier); err != nil {
		if sqlite.IsNoRows(err) {
			return errors.New("pipeline not found")
		}
		return fmt.Errorf("failed to get pipeline id: %w", err)
	}

	// Delete existing related records
	// Get pipeline_ir ID first
	query, args = db.queryBuilder.buildGetPipelineIRQuery(pipelineID)
	pipelineIRRow := db.driver.QueryRow(query, args...)

	var pipelineIRID int64
	if err := pipelineIRRow.Scan(&pipelineIRID, &pipelineID, nil, nil, nil); err != nil {
		if !sqlite.IsNoRows(err) {
			return fmt.Errorf("failed to get pipeline_ir id: %w", err)
		}
		pipelineIRID = 0
	}

	if pipelineIRID != 0 {
		// Delete old function_irs
		query, args = db.queryBuilder.buildDeleteFunctionIRsQuery(pipelineIRID)
		if err := db.driver.Exec(query, args...); err != nil {
			return fmt.Errorf("failed to delete function_irs: %w", err)
		}

		// Delete old pipe_irs
		query, args = db.queryBuilder.buildDeletePipeIRsQuery(pipelineIRID)
		if err := db.driver.Exec(query, args...); err != nil {
			return fmt.Errorf("failed to delete pipe_irs: %w", err)
		}

		// Delete old pipeline_ir
		query, args = db.queryBuilder.buildDeletePipelineIRQuery(pipelineID)
		if err := db.driver.Exec(query, args...); err != nil {
			return fmt.Errorf("failed to delete pipeline_ir: %w", err)
		}
	}

	// Insert new pipeline_ir
	query, args = db.queryBuilder.buildInsertPipelineIRQuery(
		pipelineID,
		string(record.OnCrash),
		record.OnStartup,
		record.OnTeardown,
	)
	result, err := db.driver.ExecResult(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert pipeline_ir: %w", err)
	}

	newPipelineIRID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get new pipeline_ir id: %w", err)
	}

	// Insert new function_irs
	for _, function := range record.Functions {
		parametersJSON, err := json.Marshal(function.Parameters)
		if err != nil {
			return fmt.Errorf("failed to marshal function parameters: %w", err)
		}
		returnsJSON, err := json.Marshal(function.Returns)
		if err != nil {
			return fmt.Errorf("failed to marshal function returns: %w", err)
		}

		query, args = db.queryBuilder.buildInsertFunctionIRQuery(
			newPipelineIRID,
			function.Module,
			function.Identifier,
			function.From,
			function.To,
			int(function.StartWith),
			boolToInt(function.WaitBefore),
			int(function.Maximum),
			string(parametersJSON),
			string(returnsJSON),
			function.Metadata.StaticMount,
		)
		if err := db.driver.Exec(query, args...); err != nil {
			return fmt.Errorf("failed to insert function_ir: %w", err)
		}
	}

	// Insert new pipe_irs
	for _, pipe := range record.Pipes {
		query, args = db.queryBuilder.buildInsertPipeIRQuery(
			newPipelineIRID,
			pipe.Identifier,
			int(pipe.Threshold),
			pipe.GrowthFactor,
		)
		if err := db.driver.Exec(query, args...); err != nil {
			return fmt.Errorf("failed to insert pipe_ir: %w", err)
		}
	}

	return nil
}

// Delete removes PipelineIR records from the SQLite database
func (db *SQLiteDatabase) Delete(filter database.Filter) error {

	// Ensure that only one `Get`, `Replace`, `Insert`, `Delete` operation is called at a time.
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if !db.IsConnected() {
		return database.NotConnected
	}

	// Delete by namespace and identifier - cascade will handle the rest
	query, args := db.queryBuilder.buildDeletePipelineQuery(filter.Namespace, filter.Identifier)
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

	// Return distinct namespaces
	query := db.queryBuilder.buildDistinctNamespacesQuery()
	rows, err := db.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query distinct namespaces: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var namespace string
		if err := rows.Scan(&namespace); err != nil {
			log.Printf("Error scanning distinct namespace: %v", err)
			continue
		}
		results = append(results, namespace)
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
		return fmt.Errorf("failed to ensure pipelines table exists: %w", err)
	}

	return nil
}

// Print outputs the PipelineIR records from the database to the console
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
		pipelines := db.Get(database.Filter{Namespace: namespace})
		for _, pipeline := range pipelines {
			fmt.Printf("|   ├─ %s\n", pipeline.Identifier)
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

// GetByNamespaceAndIdentifier retrieves a specific pipeline by namespace and identifier
func (db *SQLiteDatabase) GetByNamespaceAndIdentifier(namespace, identifier string) (*ScalingFunctions.PipelineIR, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if !db.IsConnected() {
		return nil, database.NotConnected
	}

	// Get pipeline ID
	query, args := db.queryBuilder.buildGetPipelineQuery(namespace, identifier)
	row := db.driver.QueryRow(query, args...)

	var pipelineID int64
	if err := row.Scan(&pipelineID, &namespace, &identifier); err != nil {
		if sqlite.IsNoRows(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan pipeline: %w", err)
	}

	// Get the pipeline IR
	return db.getPipelineIRByID(pipelineID)
}

// boolToInt converts a bool to int (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
