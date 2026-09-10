package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// T-02 records the production boundary expected from T-03:
//
//	type HealthChecker interface {
//		CheckHealth(context.Context) error
//	}
//	func NewRouter(HealthChecker) http.Handler
//
// Typed service failures expose ErrorCode() string. SERVICE_UNAVAILABLE maps
// to HTTP 503; an unclassified error maps to HTTP 500. Response bodies never
// expose the wrapped cause.

type healthCheckerStub struct {
	check func(context.Context) error
}

func (s healthCheckerStub) CheckHealth(ctx context.Context) error {
	return s.check(ctx)
}

type codedHealthError struct {
	code  string
	cause error
}

func (e codedHealthError) Error() string     { return e.cause.Error() }
func (e codedHealthError) Unwrap() error     { return e.cause }
func (e codedHealthError) ErrorCode() string { return e.code }

func TestHealthReturnsOKWhenDatabaseIsAvailable(t *testing.T) {
	database, _ := openIsolatedSQLite(t)
	router := NewRouter(healthCheckerStub{check: database.PingContext})

	response := performRequest(router, http.MethodGet, "/health")

	if response.Code != http.StatusOK {
		t.Fatalf("GET /health status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	requireJSONContentType(t, response)

	var body struct {
		Status string `json:"status"`
	}
	decodeSingleJSON(t, response, &body)
	if body.Status != "ok" {
		t.Fatalf("GET /health status body = %q, want %q", body.Status, "ok")
	}
}

func TestHealthReturnsServiceUnavailableWhenDatabaseIsUnavailable(t *testing.T) {
	database, path := openIsolatedSQLite(t)
	if err := database.Close(); err != nil {
		t.Fatalf("close isolated database: %v", err)
	}

	checker := healthCheckerStub{check: func(ctx context.Context) error {
		pingErr := database.PingContext(ctx)
		if pingErr == nil {
			return errors.New("closed database unexpectedly responded to ping")
		}
		return codedHealthError{
			code:  "SERVICE_UNAVAILABLE",
			cause: errors.New("database ping failed at " + path + ": " + pingErr.Error()),
		}
	}}
	router := NewRouter(checker)

	response := performRequest(router, http.MethodGet, "/health")

	requireErrorResponse(t, response, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE")
	requireBodyOmits(t, response, path, "database ping failed", "database is closed")
}

func TestHealthReturnsInternalErrorWithoutLeakingDetails(t *testing.T) {
	const privateDetail = "SELECT secret_value FROM internal_table: disk I/O failure"
	checker := healthCheckerStub{check: func(context.Context) error {
		return errors.New(privateDetail)
	}}
	router := NewRouter(checker)

	response := performRequest(router, http.MethodGet, "/health")

	requireErrorResponse(t, response, http.StatusInternalServerError, "INTERNAL_ERROR")
	requireBodyOmits(t, response, privateDetail, "secret_value", "internal_table", "disk I/O")
}

func TestUnknownRouteReturnsJSONNotFoundWithoutCallingHealth(t *testing.T) {
	router := NewRouter(healthCheckerStub{check: func(context.Context) error {
		t.Fatal("health checker called for unknown route")
		return nil
	}})

	response := performRequest(router, http.MethodGet, "/route-that-does-not-exist")

	requireErrorResponse(t, response, http.StatusNotFound, "ROUTE_NOT_FOUND")
}

func TestWrongMethodReturnsJSONMethodNotAllowedWithoutCallingHealth(t *testing.T) {
	router := NewRouter(healthCheckerStub{check: func(context.Context) error {
		t.Fatal("health checker called for unsupported method")
		return nil
	}})

	response := performRequest(router, http.MethodPost, "/health")

	requireErrorResponse(t, response, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

func TestSQLiteHealthFixtureUsesOneTemporaryDatabase(t *testing.T) {
	database, path := openIsolatedSQLite(t)

	if filepath.Base(path) == "library.db" {
		t.Fatalf("test database path %q must not target the runtime database", path)
	}
	if err := database.PingContext(context.Background()); err != nil {
		t.Fatalf("ping isolated database: %v", err)
	}
	if _, err := database.Exec(`CREATE TABLE isolation_marker (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create marker in isolated database: %v", err)
	}
}

func openIsolatedSQLite(t *testing.T) (*sql.DB, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "health-test.db")
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open isolated sqlite database: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})
	if err := database.PingContext(context.Background()); err != nil {
		t.Fatalf("ping isolated sqlite database: %v", err)
	}
	return database, path
}

func performRequest(router http.Handler, method, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func requireJSONContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	contentType := response.Header().Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
}

func requireErrorResponse(t *testing.T, response *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	if response.Code != wantStatus {
		t.Fatalf("response status = %d, want %d; body=%s", response.Code, wantStatus, response.Body.String())
	}
	requireJSONContentType(t, response)

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeSingleJSON(t, response, &body)
	if body.Error.Code != wantCode {
		t.Fatalf("error code = %q, want %q", body.Error.Code, wantCode)
	}
	if strings.TrimSpace(body.Error.Message) == "" {
		t.Fatal("error message must not be empty")
	}
}

func decodeSingleJSON(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()

	decoder := json.NewDecoder(strings.NewReader(response.Body.String()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("decode response JSON: %v; body=%s", err, response.Body.String())
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatalf("response must contain exactly one JSON value; trailing decode error=%v", err)
	}
}

func requireBodyOmits(t *testing.T, response *httptest.ResponseRecorder, forbidden ...string) {
	t.Helper()

	body := strings.ToLower(response.Body.String())
	for _, value := range forbidden {
		if strings.Contains(body, strings.ToLower(value)) {
			t.Errorf("response body leaked internal detail %q: %s", value, response.Body.String())
		}
	}
}
