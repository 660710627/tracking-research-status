package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

func TestResearchListServiceReturnsSharedCompleteListWithoutFiltering(t *testing.T) {
	_ = newServiceTestDatabase(t)
	parentID := int64(7)
	want := []repo.Research{
		{ID: 1, Title: "Alpha", Description: "first", Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"},
		{ID: 8, Title: "Same", Description: "continued", ContinuationOfID: &parentID, Status: "โครงการเสร็จสิ้น", Process: "การปิดบัญชีธนาคาร"},
	}
	store := researchListStoreStub{list: func(context.Context) ([]repo.Research, error) {
		return want, nil
	}}
	sut := service.NewResearchListService(store)

	firstCaller, err := sut.List(context.Background())
	if err != nil {
		t.Fatalf("first list: %v", err)
	}
	secondCaller, err := sut.List(context.Background())
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	assertServiceResearchList(t, firstCaller, want)
	assertServiceResearchList(t, secondCaller, want)
}

func TestResearchListServicePreservesNonNilEmptyList(t *testing.T) {
	_ = newServiceTestDatabase(t)
	store := researchListStoreStub{list: func(context.Context) ([]repo.Research, error) {
		return []repo.Research{}, nil
	}}
	sut := service.NewResearchListService(store)

	listed, err := sut.List(context.Background())
	if err != nil {
		t.Fatalf("list empty researches: %v", err)
	}
	if listed == nil || len(listed) != 0 {
		t.Fatalf("empty list = %#v, want non-nil empty list", listed)
	}
}

func TestResearchListServiceMapsDatabaseFailureToInternalError(t *testing.T) {
	_ = newServiceTestDatabase(t)
	store := researchListStoreStub{list: func(context.Context) ([]repo.Research, error) {
		return nil, errors.New("SQL secret: researches table")
	}}
	sut := service.NewResearchListService(store)

	_, err := sut.List(context.Background())
	if !errors.Is(err, service.ErrInternal) {
		t.Fatalf("error = %v, want ErrInternal", err)
	}
	if strings.Contains(err.Error(), "SQL secret") || strings.Contains(err.Error(), "researches table") {
		t.Fatalf("service error leaked database detail: %v", err)
	}
}

type researchListStoreStub struct {
	list func(context.Context) ([]repo.Research, error)
}

func (stub researchListStoreStub) List(ctx context.Context) ([]repo.Research, error) {
	return stub.list(ctx)
}

func assertServiceResearchList(t *testing.T, got []service.Research, want []repo.Research) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("list length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index].ID != want[index].ID || got[index].Title != want[index].Title || got[index].Description != want[index].Description || got[index].Status != want[index].Status || got[index].Process != want[index].Process {
			t.Fatalf("research[%d] = %#v, want %#v", index, got[index], want[index])
		}
		if (got[index].ContinuationOfID == nil) != (want[index].ContinuationOfID == nil) || got[index].ContinuationOfID != nil && *got[index].ContinuationOfID != *want[index].ContinuationOfID {
			t.Fatalf("research[%d].continuationOfId = %v, want %v", index, got[index].ContinuationOfID, want[index].ContinuationOfID)
		}
	}
}
