package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/db"
	"github.com/660710627/my-research/internal/handler"
	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

func listRouter(t *testing.T) (http.Handler, *repo.ResearchRepository, func()) {
	t.Helper()
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	repository := repo.NewResearchRepository(database)
	router := handler.NewRouter(service.NewHealthService(repo.NewHealthRepository(database)), handler.WithResearchLister(service.NewResearchListService(repository)))
	return router, repository, func() { _ = database.Close() }
}

func listFixture(suffix, title string) repo.CreateResearchInput {
	in := createInputForHandler(suffix)
	in.Title = title
	return in
}

func createInputForHandler(suffix string) repo.CreateResearchInput {
	return repo.CreateResearchInput{
		Title: "Research " + suffix, IsSubsidized: true, ProjectType: "RESEARCH", ResearchKind: "BUDGET", ResponsibleProjectUnit: "Project unit", ResponsibleBudgetUnit: "Budget unit", StartDate: "2024-02-29", EndDate: "2025-03-01", BudgetAmount: 12345.67, ThaiAbstract: "บทคัดย่อ", EnglishAbstract: "Abstract", Objectives: "Objectives", Keywords: "water",
		Members:  []repo.ResearchMemberInput{{FullName: "Lead", Email: "lead-" + suffix + "@example.com", Affiliation: "University", ContributionPercent: 60, Role: "LEAD"}, {FullName: "Co", Email: "co-" + suffix + "@example.com", Affiliation: "Institute", ContributionPercent: 40, Role: "CO_RESEARCHER"}},
		Contract: repo.ResearchContractInput{FundingType: "INTERNAL", FundingSourceName: "Fund", ContractNumber: "CN-" + suffix, ContractNumberKey: "cn-" + suffix, StoragePath: "private/" + suffix + ".pdf", OriginalFilename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 512},
	}
}

func get(router http.Handler, target string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, bytes.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func TestGETResearchesReturnsArrayWithContractShapeAndOrder(t *testing.T) {
	router, repository, closeDB := listRouter(t)
	defer closeDB()
	firstID, err := repository.Create(context.Background(), listFixture("first", "ข"))
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := repository.Create(context.Background(), listFixture("second", "ก"))
	if err != nil {
		t.Fatal(err)
	}
	duplicate := listFixture("duplicate", "ข")
	duplicate.ResearchKind = "CONTINUATION"
	duplicate.ContinuationOfID = &firstID
	duplicateID, err := repository.Create(context.Background(), duplicate)
	if err != nil {
		t.Fatal(err)
	}

	res := get(router, "/api/v1/researches", nil)
	if res.Code != http.StatusOK || !strings.HasPrefix(res.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("status=%d content-type=%q body=%s", res.Code, res.Header().Get("Content-Type"), res.Body.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("list=%#v", got)
	}
	ids := []int64{int64(got[0]["id"].(float64)), int64(got[1]["id"].(float64)), int64(got[2]["id"].(float64))}
	if !reflect.DeepEqual(ids, []int64{secondID, firstID, duplicateID}) {
		t.Fatalf("IDs=%v", ids)
	}
	required := []string{"id", "title", "continuationOfId", "isSubsidized", "projectMembers", "fundingType", "fundingSourceName", "contractNumber", "contractFile", "projectType", "researchKind", "responsibleProjectUnit", "responsibleBudgetUnit", "startDate", "endDate", "budgetAmount", "thaiAbstract", "englishAbstract", "objectives", "keywords", "status", "process"}
	for _, item := range got {
		for _, key := range required {
			if _, ok := item[key]; !ok {
				t.Errorf("missing %s in %#v", key, item)
			}
		}
		for _, key := range []string{"description", "binary", "storagePath", "contractNumberKey"} {
			if _, ok := item[key]; ok {
				t.Errorf("response exposes %s", key)
			}
		}
		members, ok := item["projectMembers"].([]any)
		if !ok || len(members) != 2 {
			t.Errorf("members=%#v", item["projectMembers"])
		}
		contract, ok := item["contractFile"].(map[string]any)
		if !ok || len(contract) != 3 || contract["filename"] != "contract.pdf" || contract["contentType"] != "application/pdf" || contract["sizeBytes"] != float64(512) {
			t.Errorf("contractFile=%#v", item["contractFile"])
		}
	}
}

func TestGETResearchesEmptyIsJSONArray(t *testing.T) {
	router, _, closeDB := listRouter(t)
	defer closeDB()
	res := get(router, "/api/v1/researches", nil)
	if res.Code != http.StatusOK || strings.TrimSpace(res.Body.String()) != "[]" {
		t.Fatalf("status=%d body=%q", res.Code, res.Body.String())
	}
}

func TestGETResearchesRejectsBodyAndQuery(t *testing.T) {
	for _, tc := range []struct {
		name, target string
		body         []byte
		status       int
		code         string
	}{{"body", "/api/v1/researches", []byte(`{"unexpected":true}`), 400, "INVALID_REQUEST_BODY"}, {"query", "/api/v1/researches?title=x", nil, 422, "VALIDATION_ERROR"}} {
		t.Run(tc.name, func(t *testing.T) {
			router, _, closeDB := listRouter(t)
			defer closeDB()
			postError(t, get(router, tc.target, tc.body), tc.status, tc.code)
		})
	}
}

func TestGETResearchesSanitizesDatabaseFailure(t *testing.T) {
	router, _, closeDB := listRouter(t)
	closeDB()
	postError(t, get(router, "/api/v1/researches", nil), 500, "INTERNAL_ERROR")
}
