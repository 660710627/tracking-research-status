package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/service"
)

func TestListResearchesReturnsJSONArrayWithCompleteContractFields(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	continuationID := int64(4)
	want := []service.Research{{
		ID: 9,
		ResearchData: domain.ResearchData{
			Title: "รายการงานวิจัย", ContinuationOfID: &continuationID, IsSubsidized: true,
			ProjectMembers: []domain.ProjectMember{{FullName: "หัวหน้า", Email: "lead@example.test", Affiliation: "หน่วยงาน", ContributionPercent: 60, Role: domain.MemberRoleLead}, {FullName: "ผู้ร่วม", Email: "co@example.test", Affiliation: "หน่วยงาน", ContributionPercent: 40, Role: domain.MemberRoleCoResearcher}},
			FundingType: domain.FundingTypeExternal, FundingSourceName: "แหล่งทุน", ContractNumber: "C-9", ProjectType: domain.ProjectTypeAcademicService, ResearchKind: domain.ResearchKindContinuation,
			ResponsibleProjectUnit: "หน่วยงานโครงการ", ResponsibleBudgetUnit: "หน่วยงานงบประมาณ", StartDate: "01/01/2569", EndDate: "31/12/2569", BudgetAmount: 999.99,
			ThaiAbstract: "บทคัดย่อไทย", EnglishAbstract: "English abstract", Objectives: "วัตถุประสงค์", Keywords: "คำค้น",
		},
		Contract: domain.ContractMetadata{Filename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 4}, Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ",
	}}
	router := listRouter(researchListStub{list: func(context.Context) ([]service.Research, error) { return want, nil }})

	response := performListRequest(router, "/api/v1/researches", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response)
	var got []service.Research
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode JSON array: %v; body = %s", err, response.Body.String())
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("response research = %#v, want %#v", got, want)
	}
}

func TestListResearchesReturnsEmptyJSONArray(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	router := listRouter(researchListStub{list: func(context.Context) ([]service.Research, error) { return []service.Research{}, nil }})

	response := performListRequest(router, "/api/v1/researches", "")
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" {
		t.Fatalf("response = %d %q, want 200 []", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response)
}

func TestListResearchesRejectsNonEmptyBodyAndQueryParameters(t *testing.T) {
	for _, test := range []struct{ name, target, body, code string; status int }{
		{name: "body", target: "/api/v1/researches", body: "{}", status: http.StatusBadRequest, code: "INVALID_REQUEST_BODY"},
		{name: "query", target: "/api/v1/researches?status=active", status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_ = newCreateHandlerDatabase(t)
			called := false
			router := listRouter(researchListStub{list: func(context.Context) ([]service.Research, error) { called = true; return nil, nil }})
			response := performListRequest(router, test.target, test.body)
			assertCreateError(t, response, test.status, test.code)
			if called {
				t.Fatal("service was called for a rejected request")
			}
		})
	}
}

func TestListResearchesMapsDatabaseFailureWithoutLeakingDetails(t *testing.T) {
	_ = newCreateHandlerDatabase(t)
	router := listRouter(researchListStub{list: func(context.Context) ([]service.Research, error) { return nil, errors.New("database path must not leak") }})

	response := performListRequest(router, "/api/v1/researches", "")
	assertCreateError(t, response, http.StatusInternalServerError, "INTERNAL_ERROR")
	if strings.Contains(response.Body.String(), "database path") {
		t.Fatalf("internal detail leaked: %s", response.Body.String())
	}
}

type researchListStub struct {
	list func(context.Context) ([]service.Research, error)
}

func (stub researchListStub) List(ctx context.Context) ([]service.Research, error) {
	return stub.list(ctx)
}

func listRouter(lister researchListStub) http.Handler {
	return handler.NewRouter(handler.Dependencies{ResearchList: lister})
}

func performListRequest(router http.Handler, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, strings.NewReader(body))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
