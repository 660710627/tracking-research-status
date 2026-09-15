package repo_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/660710627/my-research/internal/repo"
)

func TestListResearchesReturnsCompleteAggregatesSortedByTitleThenID(t *testing.T) {
	database, repository := createDB(t)
	first := createInput("list-first")
	first.Title = "โครงการ ข"
	firstID := mustCreate(t, repository, first)
	second := createInput("list-second")
	second.Title = "โครงการ ก"
	secondID := mustCreate(t, repository, second)
	duplicate := createInput("list-duplicate")
	duplicate.Title = first.Title
	duplicate.ResearchKind = "CONTINUATION"
	duplicate.ContinuationOfID = &firstID
	duplicateID := mustCreate(t, repository, duplicate)

	got, err := repository.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ids := []int64{got[0].ID, got[1].ID, got[2].ID}; !reflect.DeepEqual(ids, []int64{secondID, firstID, duplicateID}) {
		t.Fatalf("IDs in list order = %v", ids)
	}
	item := got[2]
	if item.Title != duplicate.Title || item.ContinuationOfID == nil || *item.ContinuationOfID != firstID || !item.IsSubsidized || item.ProjectType != duplicate.ProjectType || item.ResearchKind != duplicate.ResearchKind || item.ResponsibleProjectUnit != duplicate.ResponsibleProjectUnit || item.ResponsibleBudgetUnit != duplicate.ResponsibleBudgetUnit || item.StartDate != duplicate.StartDate || item.EndDate != duplicate.EndDate || item.BudgetAmount != duplicate.BudgetAmount || item.ThaiAbstract != duplicate.ThaiAbstract || item.EnglishAbstract != duplicate.EnglishAbstract || item.Objectives != duplicate.Objectives || item.Keywords != duplicate.Keywords || item.Status != "กำลังดำเนินการ" || item.Process != "สัญญาโครงการ" {
		t.Fatalf("incomplete research: %#v", item)
	}
	if !reflect.DeepEqual(item.Members, duplicate.Members) {
		t.Fatalf("members = %#v", item.Members)
	}
	if item.Contract.FundingType != duplicate.Contract.FundingType || item.Contract.FundingSourceName != duplicate.Contract.FundingSourceName || item.Contract.ContractNumber != duplicate.Contract.ContractNumber || item.Contract.OriginalFilename != duplicate.Contract.OriginalFilename || item.Contract.ContentType != duplicate.Contract.ContentType || item.Contract.SizeBytes != duplicate.Contract.SizeBytes {
		t.Fatalf("contract metadata = %#v", item.Contract)
	}

	typ := reflect.TypeOf(item)
	for _, forbidden := range []string{"Description", "Binary", "StoragePath", "ContractNumberKey"} {
		if _, exists := typ.FieldByName(forbidden); exists {
			t.Errorf("list record exposes %s", forbidden)
		}
	}
	contractType := reflect.TypeOf(item.Contract)
	for _, forbidden := range []string{"Binary", "StoragePath", "ContractNumberKey"} {
		if _, exists := contractType.FieldByName(forbidden); exists {
			t.Errorf("contract metadata exposes %s", forbidden)
		}
	}

	_ = database
}

func TestListResearchesEmptyAndDatabaseFailure(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, repository := createDB(t)
		got, err := repository.List(context.Background())
		if err != nil || got == nil || len(got) != 0 {
			t.Fatalf("got=%#v err=%v; want non-nil empty slice", got, err)
		}
	})
	t.Run("database_failure", func(t *testing.T) {
		database, repository := createDB(t)
		if err := database.Close(); err != nil {
			t.Fatal(err)
		}
		got, err := repository.List(context.Background())
		if got != nil || !errors.Is(err, repo.ErrInternal) {
			t.Fatalf("got=%#v err=%v; want ErrInternal", got, err)
		}
	})
}
