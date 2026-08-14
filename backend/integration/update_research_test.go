package integration_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/service"
)

func TestUpdateResearchReturnsCompleteLatestRepresentation(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	parentID := int64(4)
	want := service.Research{
		ID: 12, Title: "ชื่อใหม่", Description: "รายละเอียดใหม่", ContinuationOfID: &parentID,
		Status: "โครงการเสร็จสิ้น", Process: "บันทึกข้อตกลง",
	}
	var captured service.UpdateResearchInput
	router := handler.NewRouter(handler.Dependencies{ResearchUpdate: researchUpdaterStub{
		update: func(_ context.Context, input service.UpdateResearchInput) (service.Research, error) {
			captured = input
			return want, nil
		},
	}})

	response := performUpdateRequest(router, "/api/v1/researches/12", `{"title":"ชื่อใหม่","description":"รายละเอียดใหม่"}`, "application/json")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response)
	if captured.ID != 12 || captured.Title != "ชื่อใหม่" || captured.Description != "รายละเอียดใหม่" {
		t.Fatalf("captured input = %#v", captured)
	}
	assertResearchResponse(t, response, want)
}

func TestUpdateResearchAcceptsJSONMediaTypeCaseInsensitivelyWithParameters(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	response := performUpdateRequest(successfulUpdateRouter(), "/api/v1/researches/1", `{"title":"Title","description":"description"}`, "Application/JSON; Charset=UTF-8")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
}

func TestUpdateResearchRejectsNonPositiveOrNonIntegerPathID(t *testing.T) {
	for _, pathID := range []string{"0", "-1", "abc", "1.5"} {
		t.Run(pathID, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			called := false
			router := updateRouterWithFunc(func(context.Context, service.UpdateResearchInput) (service.Research, error) {
				called = true
				return service.Research{}, nil
			})
			response := performUpdateRequest(router, "/api/v1/researches/"+pathID, `{"title":"Title","description":"description"}`, "application/json")
			assertCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			if called {
				t.Fatal("service called for invalid path ID")
			}
		})
	}
}

func TestUpdateResearchRejectsUnsupportedMediaType(t *testing.T) {
	tests := []struct{ name, contentType string }{
		{name: "missing"},
		{name: "text", contentType: "text/plain"},
		{name: "JSON suffix", contentType: "application/problem+json"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			response := performUpdateRequest(successfulUpdateRouter(), "/api/v1/researches/1", `{"title":"Title","description":"description"}`, test.contentType)
			assertCreateError(t, response, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE")
		})
	}
}

func TestUpdateResearchRejectsInvalidJSON(t *testing.T) {
	tests := []struct{ name, body string }{
		{name: "empty"},
		{name: "whitespace", body: " \r\n\t "},
		{name: "malformed", body: `{"title":`},
		{name: "trailing data", body: `{"title":"T","description":"D"} trailing`},
		{name: "multiple values", body: `{"title":"T","description":"D"}{}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			response := performUpdateRequest(successfulUpdateRouter(), "/api/v1/researches/1", test.body, "application/json")
			assertCreateError(t, response, http.StatusBadRequest, "INVALID_JSON")
		})
	}
}

func TestUpdateResearchRejectsSchemaViolationsAndImmutableFields(t *testing.T) {
	tests := []struct{ name, body string }{
		{name: "array", body: `[]`},
		{name: "string", body: `"value"`},
		{name: "number", body: `1`},
		{name: "null", body: `null`},
		{name: "duplicate key", body: `{"title":"A","title":"B","description":"D"}`},
		{name: "unknown field", body: `{"title":"T","description":"D","extra":true}`},
		{name: "client id", body: `{"id":2,"title":"T","description":"D"}`},
		{name: "client continuation", body: `{"title":"T","description":"D","continuationOfId":null}`},
		{name: "client status", body: `{"title":"T","description":"D","status":"กำลังดำเนินการ"}`},
		{name: "client process", body: `{"title":"T","description":"D","process":"สัญญาโครงการ"}`},
		{name: "missing title", body: `{"description":"D"}`},
		{name: "missing description", body: `{"title":"T"}`},
		{name: "null title", body: `{"title":null,"description":"D"}`},
		{name: "null description", body: `{"title":"T","description":null}`},
		{name: "numeric title", body: `{"title":1,"description":"D"}`},
		{name: "boolean description", body: `{"title":"T","description":true}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			response := performUpdateRequest(successfulUpdateRouter(), "/api/v1/researches/1", test.body, "application/json")
			assertCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		})
	}
}

func TestUpdateResearchEnforcesRawBodyLimit(t *testing.T) {
	t.Run("exactly 64 KiB reaches service", func(t *testing.T) {
		_ = newCreateHandlerDatabase(t)
		body := updateBodyWithSize(t, 64*1024)
		response := performUpdateRequest(successfulUpdateRouter(), "/api/v1/researches/1", body, "application/json")
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
		}
	})
	t.Run("over 64 KiB is rejected", func(t *testing.T) {
		_ = newCreateHandlerDatabase(t)
		body := updateBodyWithSize(t, 64*1024+1)
		response := performUpdateRequest(successfulUpdateRouter(), "/api/v1/researches/1", body, "application/json")
		assertCreateError(t, response, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE")
	})
}

func TestUpdateResearchMapsTypedErrorsToContractResponses(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "not found", err: service.ErrResearchNotFound, status: http.StatusNotFound, code: "RESEARCH_NOT_FOUND"},
		{name: "title conflict", err: service.ErrTitleAlreadyExists, status: http.StatusConflict, code: "TITLE_ALREADY_EXISTS"},
		{name: "validation", err: service.ErrValidation, status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "database failure", err: errors.Join(service.ErrInternal, errors.New("secret SQL and path")), status: http.StatusInternalServerError, code: "INTERNAL_ERROR"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			router := updateRouterWithFunc(func(context.Context, service.UpdateResearchInput) (service.Research, error) {
				return service.Research{}, test.err
			})
			response := performUpdateRequest(router, "/api/v1/researches/1", `{"title":"Title","description":"description"}`, "application/json")
			assertCreateError(t, response, test.status, test.code)
			if strings.Contains(response.Body.String(), "secret SQL") {
				t.Fatalf("internal details leaked: %s", response.Body.String())
			}
		})
	}
}

type researchUpdaterStub struct {
	update func(context.Context, service.UpdateResearchInput) (service.Research, error)
}

func (stub researchUpdaterStub) Update(ctx context.Context, input service.UpdateResearchInput) (service.Research, error) {
	return stub.update(ctx, input)
}

func successfulUpdateRouter() http.Handler {
	return updateRouterWithFunc(func(_ context.Context, input service.UpdateResearchInput) (service.Research, error) {
		return service.Research{
			ID: input.ID, Title: input.Title, Description: input.Description,
			Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ",
		}, nil
	})
}

func updateRouterWithFunc(update func(context.Context, service.UpdateResearchInput) (service.Research, error)) http.Handler {
	return handler.NewRouter(handler.Dependencies{ResearchUpdate: researchUpdaterStub{update: update}})
}

func performUpdateRequest(router http.Handler, target, body, contentType string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPut, target, strings.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func updateBodyWithSize(t *testing.T, size int) string {
	t.Helper()
	prefix := `{"title":"T","description":"`
	suffix := `"}`
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
