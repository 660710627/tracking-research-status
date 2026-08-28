package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

func TestResearchListServiceReturnsCompleteRepositoryResult(t *testing.T) {
	_ = t.TempDir()
	want := []repo.Research{{
		ID: 17,
		ResearchData: domain.ResearchData{
			Title: "รายการเต็ม", IsSubsidized: true,
			ProjectMembers: []domain.ProjectMember{{FullName: "หัวหน้า", Email: "lead@example.test", Affiliation: "หน่วยงาน", ContributionPercent: 60, Role: domain.MemberRoleLead}, {FullName: "ผู้ร่วม", Email: "co@example.test", Affiliation: "หน่วยงาน", ContributionPercent: 40, Role: domain.MemberRoleCoResearcher}},
			FundingType: domain.FundingTypeExternal, FundingSourceName: "แหล่งทุน", ContractNumber: "C-17", ProjectType: domain.ProjectTypeResearch, ResearchKind: domain.ResearchKindBudget,
			ResponsibleProjectUnit: "หน่วยงานโครงการ", ResponsibleBudgetUnit: "หน่วยงานงบประมาณ", StartDate: "01/01/2569", EndDate: "31/12/2569", BudgetAmount: 123.45,
			ThaiAbstract: "บทคัดย่อไทย", EnglishAbstract: "English abstract", Objectives: "วัตถุประสงค์", Keywords: "คำค้น",
		},
		Contract: domain.ContractMetadata{Filename: "contract.pdf", ContentType: "application/pdf", SizeBytes: 4}, Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ",
	}}
	serviceUnderTest := service.NewResearchListService(researchListStoreStub{list: func(context.Context) ([]repo.Research, error) { return want, nil }})

	got, err := serviceUnderTest.List(context.Background())
	if err != nil {
		t.Fatalf("list researches: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("list result = %#v, want %#v", got, want)
	}
}

func TestResearchListServiceMapsDatabaseFailureToInternalError(t *testing.T) {
	_ = t.TempDir()
	serviceUnderTest := service.NewResearchListService(researchListStoreStub{list: func(context.Context) ([]repo.Research, error) { return nil, errors.New("database path must not leak") }})

	_, err := serviceUnderTest.List(context.Background())
	if !errors.Is(err, service.ErrInternal) {
		t.Fatalf("error = %v, want ErrInternal", err)
	}
}

type researchListStoreStub struct {
	list func(context.Context) ([]repo.Research, error)
}

func (stub researchListStoreStub) List(ctx context.Context) ([]repo.Research, error) {
	return stub.list(ctx)
}
