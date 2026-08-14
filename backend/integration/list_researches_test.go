package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/service"
)

func TestListResearchesReturnsJSONListWithExactlySixFields(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	parentID := int64(4)
	want := []service.Research{
		{ID: 1, Title: "Alpha", Description: "first", Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"},
		{ID: 5, Title: "Same", Description: "continued", ContinuationOfID: &parentID, Status: "โครงการเสร็จสิ้น", Process: "การปิดบัญชีธนาคาร"},
	}
	router := listRouter(func(context.Context) ([]service.Research, error) { return want, nil })

	response := performListRequest(router, "/api/v1/researches", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response)
	assertExactResearchArray(t, response, want)
}

func TestListResearchesReturnsSharedListToEveryCaller(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	want := []service.Research{{ID: 1, Title: "Shared", Description: "visible to all", Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"}}
	router := listRouter(func(context.Context) ([]service.Research, error) { return want, nil })

	first := performListRequest(router, "/api/v1/researches", "")
	second := performListRequest(router, "/api/v1/researches", "")
	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("caller statuses = %d/%d, want 200/200", first.Code, second.Code)
	}
	if !bytes.Equal(first.Body.Bytes(), second.Body.Bytes()) {
		t.Fatalf("callers received different lists: first=%s second=%s", first.Body.String(), second.Body.String())
	}
}

func TestListResearchesReturnsEmptyJSONArray(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	router := listRouter(func(context.Context) ([]service.Research, error) { return []service.Research{}, nil })

	response := performListRequest(router, "/api/v1/researches", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response)
	if strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("body = %q, want []", response.Body.String())
	}
}

func TestListResearchesRejectsNonEmptyRequestBody(t *testing.T) {
	for _, body := range []string{" ", "{}"} {
		t.Run("body_"+strings.ReplaceAll(body, " ", "whitespace"), func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			called := false
			router := listRouter(func(context.Context) ([]service.Research, error) {
				called = true
				return []service.Research{}, nil
			})

			response := performListRequest(router, "/api/v1/researches", body)
			assertCreateError(t, response, http.StatusBadRequest, "INVALID_REQUEST_BODY")
			if called {
				t.Fatal("list service called for request with non-empty body")
			}
		})
	}
}

func TestListResearchesRejectsEveryQueryParameter(t *testing.T) {
	for _, target := range []string{"/api/v1/researches?search=value", "/api/v1/researches?unused="} {
		t.Run(target, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			called := false
			router := listRouter(func(context.Context) ([]service.Research, error) {
				called = true
				return []service.Research{}, nil
			})

			response := performListRequest(router, target, "")
			assertCreateError(t, response, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			if called {
				t.Fatal("list service called for request with query parameters")
			}
		})
	}
}

func TestListResearchesMapsDatabaseFailureAndHidesInternalDetails(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	router := listRouter(func(context.Context) ([]service.Research, error) {
		return nil, errors.Join(service.ErrInternal, errors.New("SQL secret: researches table"))
	})

	response := performListRequest(router, "/api/v1/researches", "")
	assertCreateError(t, response, http.StatusInternalServerError, "INTERNAL_ERROR")
	if strings.Contains(response.Body.String(), "SQL secret") || strings.Contains(response.Body.String(), "researches table") {
		t.Fatalf("response leaked internal details: %s", response.Body.String())
	}
}

type researchListerStub struct {
	list func(context.Context) ([]service.Research, error)
}

func (stub researchListerStub) List(ctx context.Context) ([]service.Research, error) {
	return stub.list(ctx)
}

func listRouter(list func(context.Context) ([]service.Research, error)) http.Handler {
	return handler.NewRouter(handler.Dependencies{ResearchList: researchListerStub{list: list}})
}

func performListRequest(router http.Handler, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, strings.NewReader(body))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertExactResearchArray(t *testing.T, response *httptest.ResponseRecorder, want []service.Research) {
	t.Helper()
	var raw []map[string]json.RawMessage
	decodeSingleJSONValue(t, response, &raw)
	if len(raw) != len(want) {
		t.Fatalf("array length = %d, want %d", len(raw), len(want))
	}
	wantKeys := []string{"continuationOfId", "description", "id", "process", "status", "title"}
	for index, item := range raw {
		keys := make([]string, 0, len(item))
		for key := range item {
			keys = append(keys, key)
		}
		sortStrings(keys)
		if !reflect.DeepEqual(keys, wantKeys) {
			t.Fatalf("item %d fields = %v, want %v", index, keys, wantKeys)
		}
		encoded, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("marshal item %d: %v", index, err)
		}
		var got service.Research
		decoder := json.NewDecoder(bytes.NewReader(encoded))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&got); err != nil {
			t.Fatalf("decode item %d: %v", index, err)
		}
		if !reflect.DeepEqual(got, want[index]) {
			t.Fatalf("item %d = %#v, want %#v", index, got, want[index])
		}
	}
}

func sortStrings(values []string) {
	for index := 1; index < len(values); index++ {
		for cursor := index; cursor > 0 && values[cursor] < values[cursor-1]; cursor-- {
			values[cursor], values[cursor-1] = values[cursor-1], values[cursor]
		}
	}
}
