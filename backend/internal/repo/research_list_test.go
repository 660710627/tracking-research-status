package repo_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/660710627/my-research/internal/domain"
	"github.com/660710627/my-research/internal/repo"
)

func TestResearchRepositoryListReturnsCompleteResearchesSortedByTitleAndID(t *testing.T) {
	databaseConnection := newResearchDatabase(t)
	repository := repo.NewResearchRepository(databaseConnection)

	alpha, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("Alpha", nil))
	if err != nil {
		t.Fatalf("create Alpha: %v", err)
	}
	parent, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("Same", nil))
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("Same", &parent.ID))
	if err != nil {
		t.Fatalf("create continuation: %v", err)
	}
	zulu, err := repository.CreateMultipart(context.Background(), multipartCreateRecord("Zulu", nil))
	if err != nil {
		t.Fatalf("create Zulu: %v", err)
	}

	got, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("list researches: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("listed researches = %d, want 4", len(got))
	}
	if wantTitles := []string{"Alpha", "Same", "Same", "Zulu"}; !reflect.DeepEqual(researchTitles(got), wantTitles) {
		t.Fatalf("title order = %q, want %q", researchTitles(got), wantTitles)
	}
	if got[1].ID != parent.ID || got[2].ID != child.ID || got[1].ID >= got[2].ID {
		t.Fatalf("same-title ID order = %d, %d, want parent %d then child %d", got[1].ID, got[2].ID, parent.ID, child.ID)
	}
	for _, want := range []domain.Research{alpha, parent, child, zulu} {
		listed := researchByID(t, got, want.ID)
		if !reflect.DeepEqual(listed, want) {
			t.Fatalf("research %d = %#v, want complete persisted research %#v", want.ID, listed, want)
		}
	}
}

func TestResearchRepositoryListReturnsEmptySlice(t *testing.T) {
	repository := repo.NewResearchRepository(newResearchDatabase(t))

	got, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("list empty researches: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("list empty researches = %#v, want non-nil empty slice", got)
	}
}

func TestResearchRepositoryListReturnsDatabaseFailure(t *testing.T) {
	databaseConnection := newResearchDatabase(t)
	repository := repo.NewResearchRepository(databaseConnection)
	if err := databaseConnection.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	_, err := repository.List(context.Background())
	if err == nil {
		t.Fatal("list error = nil, want database failure")
	}
	if errors.Is(err, repo.ErrResearchNotFound) {
		t.Fatalf("database failure mapped to domain not-found error: %v", err)
	}
}

func researchTitles(researches []domain.Research) []string {
	titles := make([]string, 0, len(researches))
	for _, research := range researches {
		titles = append(titles, research.Title)
	}
	return titles
}

func researchByID(t *testing.T, researches []domain.Research, id int64) domain.Research {
	t.Helper()
	for _, research := range researches {
		if research.ID == id {
			return research
		}
	}
	t.Fatalf("research %d was not returned", id)
	return domain.Research{}
}
