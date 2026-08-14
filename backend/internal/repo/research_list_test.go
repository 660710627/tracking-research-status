package repo_test

import (
	"context"
	"testing"

	"github.com/660710627/my-research/internal/repo"
)

func TestResearchRepositoryListReturnsCompleteRecordsSortedByTitleThenID(t *testing.T) {
	database := newResearchDatabase(t)
	repository := repo.NewResearchRepository(database)

	alpha, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Alpha", Description: "first"})
	if err != nil {
		t.Fatalf("create Alpha: %v", err)
	}
	parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Same", Description: "parent"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	parentID := parent.ID
	child, err := repository.Create(context.Background(), repo.CreateResearchParams{
		Title: "Same", Description: "child", ContinuationOfID: &parentID,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	listed, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("list researches: %v", err)
	}
	if len(listed) != 3 {
		t.Fatalf("list length = %d, want 3", len(listed))
	}
	assertRepositoryResearch(t, listed[0], alpha)
	assertRepositoryResearch(t, listed[1], parent)
	assertRepositoryResearch(t, listed[2], child)
	if listed[1].Title != listed[2].Title || listed[1].ID >= listed[2].ID {
		t.Fatalf("equal titles ordered by IDs = %q/%d then %q/%d", listed[1].Title, listed[1].ID, listed[2].Title, listed[2].ID)
	}
}

func TestResearchRepositoryListReturnsNonNilEmptyCollection(t *testing.T) {
	database := newResearchDatabase(t)
	repository := repo.NewResearchRepository(database)

	listed, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("list empty researches: %v", err)
	}
	if listed == nil {
		t.Fatal("empty list = nil, want non-nil empty collection")
	}
	if len(listed) != 0 {
		t.Fatalf("empty list length = %d, want 0", len(listed))
	}
}

func TestResearchRepositoryListReturnsDatabaseFailure(t *testing.T) {
	database := newResearchDatabase(t)
	repository := repo.NewResearchRepository(database)
	if err := database.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	if _, err := repository.List(context.Background()); err == nil {
		t.Fatal("list error = nil, want database failure")
	}
}

func assertRepositoryResearch(t *testing.T, got, want repo.Research) {
	t.Helper()
	if got.ID != want.ID || got.Title != want.Title || got.Description != want.Description || got.Status != want.Status || got.Process != want.Process {
		t.Fatalf("research = %#v, want %#v", got, want)
	}
	if (got.ContinuationOfID == nil) != (want.ContinuationOfID == nil) || got.ContinuationOfID != nil && *got.ContinuationOfID != *want.ContinuationOfID {
		t.Fatalf("continuationOfId = %v, want %v", got.ContinuationOfID, want.ContinuationOfID)
	}
}
