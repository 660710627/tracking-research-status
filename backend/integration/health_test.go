package integration_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
	_ "modernc.org/sqlite"
)

func TestHealthReturnsOKWhenDatabaseIsAvailable(t *testing.T) {
	router, _ := newHealthRouter(t)

	response := performRequest(router, http.MethodGet, "/health")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertHealthResponse(t, response, "ok")
}

func TestHealthReturnsServiceUnavailableWhenDatabaseIsUnavailable(t *testing.T) {
	router, database := newHealthRouter(t)
	if err := database.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	response := performRequest(router, http.MethodGet, "/health")

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusServiceUnavailable, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertErrorResponse(t, response, "SERVICE_UNAVAILABLE")
}

func TestHealthReturnsInternalErrorForUnexpectedFailure(t *testing.T) {
	database := newTestDatabase(t)
	health := failingHealthChecker{database: database, err: errors.New("unexpected health failure")}
	router := handler.NewRouter(handler.Dependencies{Health: health})

	response := performRequest(router, http.MethodGet, "/health")

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusInternalServerError, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertErrorResponse(t, response, "INTERNAL_ERROR")
}

func TestUnknownRouteReturnsRouteNotFound(t *testing.T) {
	router, _ := newHealthRouter(t)

	response := performRequest(router, http.MethodGet, "/missing")

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusNotFound, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertErrorResponse(t, response, "ROUTE_NOT_FOUND")
}

func TestUnsupportedMethodReturnsMethodNotAllowed(t *testing.T) {
	router, _ := newHealthRouter(t)

	response := performRequest(router, http.MethodPost, "/health")

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusMethodNotAllowed, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertErrorResponse(t, response, "METHOD_NOT_ALLOWED")
}

type failingHealthChecker struct {
	database *sql.DB
	err      error
}

func (checker failingHealthChecker) Check(ctx context.Context) error {
	if err := checker.database.PingContext(ctx); err != nil {
		return err
	}
	return checker.err
}

func newHealthRouter(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	database := newTestDatabase(t)
	healthRepository := repo.NewHealthRepository(database)
	healthService := service.NewHealthService(healthRepository)
	router := handler.NewRouter(handler.Dependencies{Health: healthService})
	return router, database
}

func newTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databasePath := filepath.Join(t.TempDir(), "library.db")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	if err := database.PingContext(context.Background()); err != nil {
		_ = database.Close()
		t.Fatalf("ping SQLite database: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})
	return database
}

func performRequest(router http.Handler, method, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertJSONContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	mediaType, _, err := mime.ParseMediaType(response.Header().Get("Content-Type"))
	if err != nil {
		t.Fatalf("parse Content-Type: %v", err)
	}
	if mediaType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", response.Header().Get("Content-Type"))
	}
}

func assertHealthResponse(t *testing.T, response *httptest.ResponseRecorder, wantStatus string) {
	t.Helper()
	var body struct {
		Status string `json:"status"`
	}
	decodeSingleJSONValue(t, response, &body)
	if body.Status != wantStatus {
		t.Fatalf("body.status = %q, want %q", body.Status, wantStatus)
	}
}

func assertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeSingleJSONValue(t, response, &body)
	if body.Error.Code != wantCode {
		t.Fatalf("error.code = %q, want %q", body.Error.Code, wantCode)
	}
	if body.Error.Message == "" {
		t.Fatal("error.message must not be empty")
	}
}

func decodeSingleJSONValue(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		t.Fatalf("decode response body: %v; body = %s", err, response.Body.String())
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatalf("response must contain exactly one JSON value; trailing decode error = %v", err)
	}
}
