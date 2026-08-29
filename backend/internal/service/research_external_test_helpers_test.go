package service_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newServiceTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "service-external.db")
	databaseConnection, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })
	return databaseConnection
}
