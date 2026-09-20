package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/statistic"
)

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
