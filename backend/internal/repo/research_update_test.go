package repo_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/660710627/my-research/internal/repo"
)

func TestResearchRepositoryUpdateReplacesDetailsAndPreservesImmutableFields(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Parent", Description: "parent"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	parentID := parent.ID
	child, err := repository.Create(context.Background(), repo.CreateResearchParams{
		Title: "Child", Description: "old description", ContinuationOfID: &parentID,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE researches SET status = ?, process = ? WHERE id = ?`,
		"โครงการเสร็จสิ้น", "บันทึกข้อตกลง", child.ID); err != nil {
		t.Fatalf("prepare persisted state: %v", err)
	}

	updated, err := repository.Update(context.Background(), repo.UpdateResearchParams{
		ID: child.ID, Title: "New child", Description: "new description",
	})
	if err != nil {
		t.Fatalf("update research: %v", err)
	}
	if updated.ID != child.ID || updated.Title != "New child" || updated.Description != "new description" {
		t.Fatalf("updated = %#v, want replaced details and original ID", updated)
	}
	if updated.ContinuationOfID == nil || *updated.ContinuationOfID != parentID {
		t.Fatalf("continuationOfId = %v, want %d", updated.ContinuationOfID, parentID)
	}
	if updated.Status != "โครงการเสร็จสิ้น" || updated.Process != "บันทึกข้อตกลง" {
		t.Fatalf("status/process = %q/%q, want persisted values", updated.Status, updated.Process)
	}

	var persisted repo.Research
	var continuationID int64
	err = db.QueryRowContext(context.Background(), `
		SELECT id, title, description, continuation_of_id, status, process
		FROM researches WHERE id = ?`, child.ID,
	).Scan(&persisted.ID, &persisted.Title, &persisted.Description, &continuationID, &persisted.Status, &persisted.Process)
	if err != nil {
		t.Fatalf("read updated research: %v", err)
	}
	if persisted.ID != updated.ID || persisted.Title != updated.Title || persisted.Description != updated.Description ||
		continuationID != parentID || persisted.Status != updated.Status || persisted.Process != updated.Process {
		t.Fatalf("persisted row does not match update result: %#v continuation=%d", persisted, continuationID)
	}
}

func TestResearchRepositoryUpdateEnforcesRootAndContinuationTitleRules(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *repo.ResearchRepository)
	}{
		{
			name: "root retains title used by continuation",
			run: func(t *testing.T, repository *repo.ResearchRepository) {
				root, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Shared", Description: "root"})
				if err != nil {
					t.Fatalf("create root: %v", err)
				}
				rootID := root.ID
				if _, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Shared", Description: "child", ContinuationOfID: &rootID}); err != nil {
					t.Fatalf("create continuation: %v", err)
				}
				updated, err := repository.Update(context.Background(), repo.UpdateResearchParams{ID: root.ID, Title: "Shared", Description: "changed"})
				if err != nil || updated.Description != "changed" {
					t.Fatalf("update root description: %#v, %v", updated, err)
				}
			},
		},
		{
			name: "root cannot change to any existing title",
			run: func(t *testing.T, repository *repo.ResearchRepository) {
				root, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Root", Description: "original"})
				if err != nil {
					t.Fatalf("create root: %v", err)
				}
				parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Parent", Description: "parent"})
				if err != nil {
					t.Fatalf("create parent: %v", err)
				}
				parentID := parent.ID
				if _, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Reserved", Description: "child", ContinuationOfID: &parentID}); err != nil {
					t.Fatalf("create continuation: %v", err)
				}
				_, err = repository.Update(context.Background(), repo.UpdateResearchParams{ID: root.ID, Title: "Reserved", Description: "must roll back"})
				if !errors.Is(err, repo.ErrTitleAlreadyExists) {
					t.Fatalf("error = %v, want ErrTitleAlreadyExists", err)
				}
				var title, description string
				if err := repositoryTestQuery(root.ID, repository, &title, &description); err != nil {
					t.Fatal(err)
				}
				if title != "Root" || description != "original" {
					t.Fatalf("row partially changed to %q/%q", title, description)
				}
			},
		},
		{
			name: "continuation may change to duplicate title",
			run: func(t *testing.T, repository *repo.ResearchRepository) {
				root, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Root", Description: "root"})
				if err != nil {
					t.Fatalf("create root: %v", err)
				}
				rootID := root.ID
				child, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Child", Description: "child", ContinuationOfID: &rootID})
				if err != nil {
					t.Fatalf("create child: %v", err)
				}
				updated, err := repository.Update(context.Background(), repo.UpdateResearchParams{ID: child.ID, Title: "Root", Description: "changed"})
				if err != nil || updated.Title != "Root" {
					t.Fatalf("update continuation: %#v, %v", updated, err)
				}
			},
		},
		{
			name: "root comparison is case sensitive",
			run: func(t *testing.T, repository *repo.ResearchRepository) {
				root, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Original", Description: "root"})
				if err != nil {
					t.Fatalf("create root: %v", err)
				}
				if _, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Reserved", Description: "other"}); err != nil {
					t.Fatalf("create other: %v", err)
				}
				if _, err := repository.Update(context.Background(), repo.UpdateResearchParams{ID: root.ID, Title: "reserved", Description: "changed"}); err != nil {
					t.Fatalf("case-sensitive update rejected: %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newResearchDatabase(t)
			test.run(t, repo.NewResearchRepository(db))
		})
	}
}

func TestResearchRepositoryUpdateReturnsNotFound(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	_, err := repository.Update(context.Background(), repo.UpdateResearchParams{ID: 999, Title: "Title", Description: "description"})
	if !errors.Is(err, repo.ErrResearchNotFound) {
		t.Fatalf("error = %v, want ErrResearchNotFound", err)
	}
}

func TestResearchRepositoryUpdateRollsBackWhenEitherFieldIsInvalid(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	created, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Original", Description: "original description"})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}

	if _, err := repository.Update(context.Background(), repo.UpdateResearchParams{ID: created.ID, Title: "Changed", Description: "bad\rdescription"}); err == nil {
		t.Fatal("invalid update succeeded")
	}
	var title, description string
	if err := repositoryTestQuery(created.ID, repository, &title, &description); err != nil {
		t.Fatal(err)
	}
	if title != "Original" || description != "original description" {
		t.Fatalf("update was not atomic: title=%q description=%q", title, description)
	}
}

func TestResearchRepositoryConcurrentRootTitleConflictAllowsExactlyOne(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	first, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "First", Description: "first"})
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	second, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Second", Description: "second"})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, id := range []int64{first.ID, second.ID} {
		wait.Add(1)
		go func(id int64) {
			defer wait.Done()
			<-start
			_, err := repository.Update(context.Background(), repo.UpdateResearchParams{ID: id, Title: "Same", Description: "updated"})
			results <- err
		}(id)
	}
	close(start)
	wait.Wait()
	close(results)

	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, repo.ErrTitleAlreadyExists):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent result: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes/conflicts = %d/%d, want 1/1", successes, conflicts)
	}
}

func TestResearchRepositoryUpdateReturnsDatabaseFailure(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	if _, err := repository.Update(context.Background(), repo.UpdateResearchParams{ID: 1, Title: "Title", Description: "description"}); err == nil {
		t.Fatal("update against closed database succeeded")
	}
}

// repositoryTestQuery intentionally reads through the repository's test database
// using List, so the test does not depend on private repository fields.
func repositoryTestQuery(id int64, repository *repo.ResearchRepository, title, description *string) error {
	items, err := repository.List(context.Background())
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.ID == id {
			*title, *description = item.Title, item.Description
			return nil
		}
	}
	return errors.New("research row not found")
}
