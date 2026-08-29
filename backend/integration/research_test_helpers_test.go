package integration_test

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/660710627/my-research/internal/service"
	_ "modernc.org/sqlite"
)

func newCreateHandlerDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "handler.db")
	databaseConnection, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := databaseConnection.PingContext(context.Background()); err != nil {
		_ = databaseConnection.Close()
		t.Fatalf("ping SQLite: %v", err)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })
	return databaseConnection
}

func assertCreateError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, status, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertErrorResponse(t, response, code)
}

type researchProcessUpdaterStub struct {
	update func(context.Context, service.UpdateResearchProcessInput) (service.Research, error)
}

func (stub researchProcessUpdaterStub) UpdateProcess(ctx context.Context, input service.UpdateResearchProcessInput) (service.Research, error) {
	return stub.update(ctx, input)
}
