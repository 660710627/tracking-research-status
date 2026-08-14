package repo_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/660710627/my-research/internal/repo"
)

func TestResearchRepositoryDeleteRemovesResearchFromList(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	created, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Delete me", Description: "description"})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	if err := repository.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete research: %v", err)
	}
	assertResearchCount(t, db, 0)
	items, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("list researches: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("list after delete = %#v, want empty", items)
	}
}

func TestResearchRepositoryDeleteReturnsNotFoundForMissingOrRepeatedID(t *testing.T) {
	t.Run("missing ID", func(t *testing.T) {
		db := newResearchDatabase(t)
		repository := repo.NewResearchRepository(db)
		if err := repository.Delete(context.Background(), 999); !errors.Is(err, repo.ErrResearchNotFound) {
			t.Fatalf("error = %v, want ErrResearchNotFound", err)
		}
	})

	t.Run("repeated delete", func(t *testing.T) {
		db := newResearchDatabase(t)
		repository := repo.NewResearchRepository(db)
		created, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Delete once", Description: "description"})
		if err != nil {
			t.Fatalf("create research: %v", err)
		}
		if err := repository.Delete(context.Background(), created.ID); err != nil {
			t.Fatalf("first delete: %v", err)
		}
		if err := repository.Delete(context.Background(), created.ID); !errors.Is(err, repo.ErrResearchNotFound) {
			t.Fatalf("second delete error = %v, want ErrResearchNotFound", err)
		}
	})
}

func TestResearchRepositoryDeleteRejectsParentWithContinuations(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Parent", Description: "parent"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	parentID := parent.ID
	child, err := repository.Create(context.Background(), repo.CreateResearchParams{
		Title: "Child", Description: "child", ContinuationOfID: &parentID,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	if err := repository.Delete(context.Background(), parent.ID); !errors.Is(err, repo.ErrResearchHasContinuations) {
		t.Fatalf("error = %v, want ErrResearchHasContinuations", err)
	}
	assertResearchCount(t, db, 2)
	items, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("list after rejected delete: %v", err)
	}
	seen := map[int64]bool{}
	for _, item := range items {
		seen[item.ID] = true
	}
	if !seen[parent.ID] || !seen[child.ID] {
		t.Fatalf("parent or continuation disappeared after rejected delete: %#v", items)
	}
}

func TestResearchDatabaseRestrictsDeletingReferencedParentDirectly(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Parent", Description: "parent"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	parentID := parent.ID
	if _, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Child", Description: "child", ContinuationOfID: &parentID}); err != nil {
		t.Fatalf("create child: %v", err)
	}

	if _, err := db.ExecContext(context.Background(), `DELETE FROM researches WHERE id = ?`, parent.ID); err == nil {
		t.Fatal("database allowed deleting a referenced parent")
	}
	assertResearchCount(t, db, 2)
}

func TestResearchRepositoryDeleteNeverReusesDeletedID(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	first, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "First", Description: "first"})
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	if err := repository.Delete(context.Background(), first.ID); err != nil {
		t.Fatalf("delete first: %v", err)
	}
	second, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Second", Description: "second"})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}
	if second.ID <= first.ID {
		t.Fatalf("new ID = %d, want greater than deleted ID %d", second.ID, first.ID)
	}
}

func TestResearchRepositoryConcurrentDeleteSucceedsExactlyOnce(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	created, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Concurrent delete", Description: "description"})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			results <- repository.Delete(context.Background(), created.ID)
		}()
	}
	close(start)
	wait.Wait()
	close(results)

	successes, notFound := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, repo.ErrResearchNotFound):
			notFound++
		default:
			t.Fatalf("unexpected concurrent delete error: %v", err)
		}
	}
	if successes != 1 || notFound != 1 {
		t.Fatalf("success/not-found = %d/%d, want 1/1", successes, notFound)
	}
	assertResearchCount(t, db, 0)
}

func TestResearchRepositoryDeleteReturnsDatabaseFailure(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	if err := repository.Delete(context.Background(), 1); err == nil {
		t.Fatal("delete against closed database succeeded")
	}
}
