package service

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/660710627/my-research/internal/repo"
	_ "modernc.org/sqlite"
)

type researchProcessStoreStub struct {
	update func(context.Context, repo.UpdateResearchProcessParams) (repo.Research, error)
}

func (stub researchProcessStoreStub) UpdateProcess(ctx context.Context, params repo.UpdateResearchProcessParams) (repo.Research, error) {
	return stub.update(ctx, params)
}

func newServiceTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "service.db")
	databaseConnection, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })
	return databaseConnection
}
