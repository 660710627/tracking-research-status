package repo_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	database "github.com/660710627/my-research/internal/db"
	_ "modernc.org/sqlite"
)

const (
	initialStatus  = "กำลังดำเนินการ"
	initialProcess = "สัญญาโครงการ"
)

func newResearchDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "research.db")
	databaseConnection, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	databaseConnection.SetMaxOpenConns(12)
	if err := database.Initialize(context.Background(), databaseConnection); err != nil {
		_ = databaseConnection.Close()
		t.Fatalf("initialize schema: %v", err)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })
	return databaseConnection
}

func assertResearchCount(t *testing.T, databaseConnection *sql.DB, want int) {
	t.Helper()
	var count int
	if err := databaseConnection.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM researches").Scan(&count); err != nil {
		t.Fatalf("count researches: %v", err)
	}
	if count != want {
		t.Fatalf("research count = %d, want %d", count, want)
	}
}
