package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

type listRepositoryStub struct {
	items []repo.Research
	err   error
}

func (s listRepositoryStub) List(context.Context) ([]repo.Research, error) { return s.items, s.err }

func TestResearchListServiceFormatsDatesAndPreservesCompleteData(t *testing.T) {
	parent := int64(7)
	record := repo.Research{
		ID: 8, Title: "โครงการต่อเนื่อง", ContinuationOfID: &parent, IsSubsidized: true,
		ProjectType: "RESEARCH", ResearchKind: "CONTINUATION", ResponsibleProjectUnit: "หน่วยโครงการ", ResponsibleBudgetUnit: "หน่วยงบ",
		StartDate: "2024-02-29", EndDate: "2025-03-01", BudgetAmount: 100.25, ThaiAbstract: "ไทย", EnglishAbstract: "English", Objectives: "เป้าหมาย", Keywords: "คำค้น", Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ",
		Members:  []repo.ResearchMember{{FullName: "Lead", Email: "lead@example.com", Affiliation: "University", ContributionPercent: 60, Role: "LEAD"}, {FullName: "Co", Email: "co@example.com", Affiliation: "Institute", ContributionPercent: 40, Role: "CO_RESEARCHER"}},
		Contract: repo.ResearchContract{FundingType: "INTERNAL", FundingSourceName: "Fund", ContractNumber: "CN-8", OriginalFilename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 512},
	}
	got, err := service.NewResearchListService(listRepositoryStub{items: []repo.Research{record}}).List(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	if got[0].StartDate != "29/02/2567" || got[0].EndDate != "01/03/2568" || got[0].ID != record.ID || len(got[0].ProjectMembers) != 2 || got[0].ContractFile.Filename != "contract.pdf" || got[0].ContractFile.ContentType != "application/pdf" || got[0].ContractFile.SizeBytes != 512 {
		t.Fatalf("mapped research = %#v", got[0])
	}
}

func TestResearchListServiceSanitizesRepositoryFailure(t *testing.T) {
	_, err := service.NewResearchListService(listRepositoryStub{err: errors.New("injected-secret database failure")}).List(context.Background())
	var coded interface{ ErrorCode() string }
	if !errors.As(err, &coded) || coded.ErrorCode() != "INTERNAL_ERROR" || err.Error() == "injected-secret database failure" {
		t.Fatalf("error = %v", err)
	}
}
