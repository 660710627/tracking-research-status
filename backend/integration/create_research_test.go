package integration_test

import (
	"bytes"
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

	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/service"
	_ "modernc.org/sqlite"
)

func TestCreateResearchReturnsCompletePersistedRepresentation(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	parentID := int64(7)
	want := service.Research{
		ID: 12, Title: "งานต่อเนื่อง", Description: "รายละเอียด", ContinuationOfID: &parentID,
		Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ",
	}
	var captured service.CreateResearchInput
	router := handler.NewRouter(handler.Dependencies{Researches: researchCreatorStub{
		create: func(_ context.Context, input service.CreateResearchInput) (service.Research, error) {
			captured = input
			return want, nil
		},
	}})

	response := performCreateRequest(router, `{"title":"งานต่อเนื่อง","description":"รายละเอียด","continuationOfId":7}`, "application/json")
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response)
	if captured.Title != "งานต่อเนื่อง" || captured.Description != "รายละเอียด" || captured.ContinuationOfID == nil || *captured.ContinuationOfID != parentID {
		t.Fatalf("captured input = %#v", captured)
	}
	assertResearchResponse(t, response, want)
}

func TestCreateResearchAcceptsJSONMediaTypeCaseInsensitivelyWithParameters(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	router := successfulCreateRouter()
	response := performCreateRequest(router, `{"title":"Title","description":"description","continuationOfId":null}`, "Application/JSON; Charset=UTF-8")
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
}

func TestCreateResearchRejectsUnsupportedMediaType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{name: "missing"},
		{name: "text", contentType: "text/plain"},
		{name: "JSON suffix is not exact media type", contentType: "application/problem+json"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			response := performCreateRequest(successfulCreateRouter(), `{"title":"Title","description":"description","continuationOfId":null}`, test.contentType)
			assertCreateError(t, response, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE")
		})
	}
}

func TestCreateResearchRejectsInvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty"},
		{name: "whitespace", body: " \r\n\t "},
		{name: "malformed", body: `{"title":`},
		{name: "trailing data", body: `{"title":"T","description":"D","continuationOfId":null} trailing`},
		{name: "multiple values", body: `{"title":"T","description":"D","continuationOfId":null}{}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			response := performCreateRequest(successfulCreateRouter(), test.body, "application/json")
			assertCreateError(t, response, http.StatusBadRequest, "INVALID_JSON")
		})
	}
}

func TestCreateResearchRejectsNonObjectDuplicateUnknownOrMissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "array", body: `[]`},
		{name: "string", body: `"value"`},
		{name: "number", body: `1`},
		{name: "null", body: `null`},
		{name: "duplicate key", body: `{"title":"A","title":"B","description":"D","continuationOfId":null}`},
		{name: "unknown field", body: `{"title":"T","description":"D","continuationOfId":null,"extra":true}`},
		{name: "client id", body: `{"id":1,"title":"T","description":"D","continuationOfId":null}`},
		{name: "client status", body: `{"title":"T","description":"D","continuationOfId":null,"status":"กำลังดำเนินการ"}`},
		{name: "client process", body: `{"title":"T","description":"D","continuationOfId":null,"process":"สัญญาโครงการ"}`},
		{name: "missing title", body: `{"description":"D","continuationOfId":null}`},
		{name: "missing description", body: `{"title":"T","continuationOfId":null}`},
		{name: "missing continuation", body: `{"title":"T","description":"D"}`},
		{name: "null title", body: `{"title":null,"description":"D","continuationOfId":null}`},
		{name: "null description", body: `{"title":"T","description":null,"continuationOfId":null}`},
		{name: "string continuation", body: `{"title":"T","description":"D","continuationOfId":"1"}`},
		{name: "fractional continuation", body: `{"title":"T","description":"D","continuationOfId":1.5}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			response := performCreateRequest(successfulCreateRouter(), test.body, "application/json")
			assertCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		})
	}
}

func TestCreateResearchEnforcesRawBodyLimit(t *testing.T) {
	t.Run("exactly 64 KiB is accepted by handler", func(t *testing.T) {
		_ = newCreateHandlerDatabase(t)
		body := createBodyWithSize(t, 64*1024)
		response := performCreateRequest(successfulCreateRouter(), body, "application/json")
		if response.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201 at exact limit; body = %s", response.Code, response.Body.String())
		}
	})
	t.Run("over 64 KiB is rejected", func(t *testing.T) {
		_ = newCreateHandlerDatabase(t)
		body := createBodyWithSize(t, 64*1024+1)
		response := performCreateRequest(successfulCreateRouter(), body, "application/json")
		assertCreateError(t, response, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE")
	})
}

func TestCreateResearchMapsTypedErrorsAndHidesInternalDetails(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		status     int
		code       string
	}{
		{name: "continuation missing", serviceErr: service.ErrContinuationNotFound, status: http.StatusNotFound, code: "CONTINUATION_NOT_FOUND"},
		{name: "title conflict", serviceErr: service.ErrTitleAlreadyExists, status: http.StatusConflict, code: "TITLE_ALREADY_EXISTS"},
		{name: "validation", serviceErr: service.ErrValidation, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "database failure", serviceErr: fmtInternalError("SQL secret: researches"), status: http.StatusInternalServerError, code: "INTERNAL_ERROR"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			router := handler.NewRouter(handler.Dependencies{Researches: researchCreatorStub{
				create: func(context.Context, service.CreateResearchInput) (service.Research, error) {
					return service.Research{}, test.serviceErr
				},
			}})
			response := performCreateRequest(router, `{"title":"Title","description":"description","continuationOfId":null}`, "application/json")
			assertCreateError(t, response, test.status, test.code)
			if strings.Contains(response.Body.String(), "SQL secret") || strings.Contains(response.Body.String(), "researches") {
				t.Fatalf("response leaked internal details: %s", response.Body.String())
			}
		})
	}
}

type researchCreatorStub struct {
	create func(context.Context, service.CreateResearchInput) (service.Research, error)
}

func (stub researchCreatorStub) Create(ctx context.Context, input service.CreateResearchInput) (service.Research, error) {
	return stub.create(ctx, input)
}

func successfulCreateRouter() http.Handler {
	return handler.NewRouter(handler.Dependencies{Researches: researchCreatorStub{
		create: func(_ context.Context, input service.CreateResearchInput) (service.Research, error) {
			return service.Research{
				ID: 1, Title: input.Title, Description: input.Description,
				ContinuationOfID: input.ContinuationOfID,
				Status:           "กำลังดำเนินการ", Process: "สัญญาโครงการ",
			}, nil
		},
	}})
}

func performCreateRequest(router http.Handler, body, contentType string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/researches", strings.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertCreateError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, status, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertErrorResponse(t, response, code)
}

func assertResearchResponse(t *testing.T, response *httptest.ResponseRecorder, want service.Research) {
	t.Helper()
	var body struct {
		ID               int64  `json:"id"`
		Title            string `json:"title"`
		Description      string `json:"description"`
		ContinuationOfID *int64 `json:"continuationOfId"`
		Status           string `json:"status"`
		Process          string `json:"process"`
	}
	decoder := json.NewDecoder(bytes.NewReader(response.Body.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		t.Fatalf("decode research response: %v; body = %s", err, response.Body.String())
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatalf("response must contain one JSON value; trailing error = %v", err)
	}
	if body.ID != want.ID || body.Title != want.Title || body.Description != want.Description || body.Status != want.Status || body.Process != want.Process {
		t.Fatalf("response = %#v, want %#v", body, want)
	}
	if (body.ContinuationOfID == nil) != (want.ContinuationOfID == nil) || body.ContinuationOfID != nil && *body.ContinuationOfID != *want.ContinuationOfID {
		t.Fatalf("continuationOfId = %v, want %v", body.ContinuationOfID, want.ContinuationOfID)
	}
}

func createBodyWithSize(t *testing.T, size int) string {
	t.Helper()
	prefix := `{"title":"T","description":"`
	suffix := `","continuationOfId":null}`
	padding := size - len(prefix) - len(suffix)
	if padding < 0 {
		t.Fatalf("requested body size %d is too small", size)
	}
	body := prefix + strings.Repeat("a", padding) + suffix
	if len(body) != size {
		t.Fatalf("body size = %d, want %d", len(body), size)
	}
	return body
}

func newCreateHandlerDatabase(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "handler.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Fatalf("ping SQLite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func fmtInternalError(detail string) error {
	return errors.Join(service.ErrInternal, errors.New(detail))
}
