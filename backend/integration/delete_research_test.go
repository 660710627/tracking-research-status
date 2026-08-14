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

func TestDeleteResearchReturnsNoContentWithEmptyBody(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	var capturedID int64
	router := deleteRouterWithFunc(func(_ context.Context, id int64) error {
		capturedID = id
		return nil
	})

	response := performDeleteRequest(router, "/api/v1/researches/42", "")
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body = %s", response.Code, response.Body.String())
	}
	if response.Body.Len() != 0 {
		t.Fatalf("204 response body = %q, want empty", response.Body.String())
	}
	if capturedID != 42 {
		t.Fatalf("service ID = %d, want 42", capturedID)
	}
}

func TestDeleteResearchRejectsNonPositiveOrNonIntegerPathID(t *testing.T) {
	for _, pathID := range []string{"0", "-1", "abc", "1.5"} {
		t.Run(pathID, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			called := false
			router := deleteRouterWithFunc(func(context.Context, int64) error {
				called = true
				return nil
			})
			response := performDeleteRequest(router, "/api/v1/researches/"+pathID, "")
			assertCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			if called {
				t.Fatal("service called for invalid path ID")
			}
		})
	}
}

func TestDeleteResearchRejectsAnyNonEmptyBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "JSON object", body: `{}`},
		{name: "whitespace", body: " \r\n\t"},
		{name: "single byte", body: "x"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			called := false
			router := deleteRouterWithFunc(func(context.Context, int64) error {
				called = true
				return nil
			})
			response := performDeleteRequest(router, "/api/v1/researches/1", test.body)
			assertCreateError(t, response, http.StatusBadRequest, "INVALID_REQUEST_BODY")
			if called {
				t.Fatal("service called for request with body")
			}
		})
	}
}

func TestDeleteResearchRejectsQueryParameters(t *testing.T) {
	tests := []string{"?force=true", "?unknown=", "?id=1"}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			called := false
			router := deleteRouterWithFunc(func(context.Context, int64) error {
				called = true
				return nil
			})
			response := performDeleteRequest(router, "/api/v1/researches/1"+query, "")
			assertCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			if called {
				t.Fatal("service called for request with query parameters")
			}
		})
	}
}

func TestDeleteResearchRepeatedRequestReturnsNotFound(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	callCount := 0
	router := deleteRouterWithFunc(func(context.Context, int64) error {
		callCount++
		if callCount > 1 {
			return service.ErrResearchNotFound
		}
		return nil
	})

	first := performDeleteRequest(router, "/api/v1/researches/1", "")
	if first.Code != http.StatusNoContent || first.Body.Len() != 0 {
		t.Fatalf("first response = %d %q, want 204 empty", first.Code, first.Body.String())
	}
	second := performDeleteRequest(router, "/api/v1/researches/1", "")
	assertCreateError(t, second, http.StatusNotFound, "RESEARCH_NOT_FOUND")
}

func TestDeleteResearchMapsTypedErrorsToContractResponses(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "not found", err: service.ErrResearchNotFound, status: http.StatusNotFound, code: "RESEARCH_NOT_FOUND"},
		{name: "has continuations", err: service.ErrResearchHasContinuations, status: http.StatusConflict, code: "RESEARCH_HAS_CONTINUATIONS"},
		{name: "database failure", err: errors.Join(service.ErrInternal, errors.New("secret SQL and path")), status: http.StatusInternalServerError, code: "INTERNAL_ERROR"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			router := deleteRouterWithFunc(func(context.Context, int64) error { return test.err })
			response := performDeleteRequest(router, "/api/v1/researches/1", "")
			assertCreateError(t, response, test.status, test.code)
			if strings.Contains(response.Body.String(), "secret SQL") {
				t.Fatalf("internal details leaked: %s", response.Body.String())
			}
		})
	}
}

func TestDeleteResearchReturnsInternalErrorWhenDependencyIsMissing(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	response := performDeleteRequest(handler.NewRouter(handler.Dependencies{}), "/api/v1/researches/1", "")
	assertCreateError(t, response, http.StatusInternalServerError, "INTERNAL_ERROR")
}

type researchDeleterStub struct {
	delete func(context.Context, int64) error
}

func (stub researchDeleterStub) Delete(ctx context.Context, id int64) error {
	return stub.delete(ctx, id)
}

func deleteRouterWithFunc(deleteResearch func(context.Context, int64) error) http.Handler {
	return handler.NewRouter(handler.Dependencies{ResearchDelete: researchDeleterStub{delete: deleteResearch}})
}

func performDeleteRequest(router http.Handler, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodDelete, target, strings.NewReader(body))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
