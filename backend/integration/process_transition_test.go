package integration_test

import (
	"context"
	"errors"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateResearchProcessValidatesContractAndMapsErrors(t *testing.T) {
	for _, test := range []struct {
		body, contentType string
		want              int
	}{{``, "application/json", 400}, {`[]`, "application/json", 422}, {`{"process":"other"}`, "application/json", 422}, {`{"process":"บันทึกข้อตกลง","extra":true}`, "application/json", 422}, {`{"process":"บันทึกข้อตกลง"}`, "text/plain", 415}} {
		response := performProcessRequest(processRouter(func(context.Context, service.UpdateResearchProcessInput) (service.Research, error) {
			return service.Research{}, nil
		}), "/api/v1/researches/1/process", test.body, test.contentType)
		if response.Code != test.want {
			t.Fatalf("status=%d want=%d body=%s", response.Code, test.want, response.Body.String())
		}
	}
	for _, test := range []struct {
		err  error
		want int
	}{{service.ErrResearchNotFound, 404}, {service.ErrInvalidProcessTransition, 409}, {service.ErrProjectAlreadyEnded, 409}, {errors.Join(service.ErrInternal, errors.New("SQL secret")), 500}} {
		response := performProcessRequest(processRouter(func(context.Context, service.UpdateResearchProcessInput) (service.Research, error) {
			return service.Research{}, test.err
		}), "/api/v1/researches/1/process", `{"process":"บันทึกข้อตกลง"}`, "application/json")
		if response.Code != test.want || strings.Contains(response.Body.String(), "SQL secret") {
			t.Fatalf("status/body=%d/%s", response.Code, response.Body.String())
		}
	}
}

func processRouter(update func(context.Context, service.UpdateResearchProcessInput) (service.Research, error)) http.Handler {
	return handler.NewRouter(handler.Dependencies{ResearchProcess: researchProcessUpdaterStub{update: update}})
}
func performProcessRequest(router http.Handler, path, body, contentType string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPatch, path, strings.NewReader(body))
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
