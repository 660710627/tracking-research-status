package repo_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	database "github.com/660710627/my-research/internal/db"
	"github.com/660710627/my-research/internal/repo"
	_ "modernc.org/sqlite"
)

const (
	initialStatus  = "กำลังดำเนินการ"
	initialProcess = "สัญญาโครงการ"
)

func TestResearchRepositoryCreatePersistsCompleteRootAtomically(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)

	created, err := repository.Create(context.Background(), repo.CreateResearchParams{
		Title:       "โครงการหลัก",
		Description: "รายละเอียดโครงการ",
	})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("id = %d, want positive", created.ID)
	}
	if created.Title != "โครงการหลัก" || created.Description != "รายละเอียดโครงการ" {
		t.Fatalf("created = %#v, want persisted title and description", created)
	}
	if created.ContinuationOfID != nil {
		t.Fatalf("continuationOfId = %v, want nil", created.ContinuationOfID)
	}
	if created.Status != initialStatus || created.Process != initialProcess {
		t.Fatalf("status/process = %q/%q, want %q/%q", created.Status, created.Process, initialStatus, initialProcess)
	}

	var title, description, status, process string
	var continuation sql.NullInt64
	err = db.QueryRowContext(context.Background(), `
		SELECT title, description, continuation_of_id, status, process
		FROM researches WHERE id = ?`, created.ID,
	).Scan(&title, &description, &continuation, &status, &process)
	if err != nil {
		t.Fatalf("query persisted research: %v", err)
	}
	if title != created.Title || description != created.Description || continuation.Valid || status != created.Status || process != created.Process {
		t.Fatalf("database row does not match response: title=%q description=%q continuation=%v status=%q process=%q", title, description, continuation, status, process)
	}
}

func TestResearchRepositorySupportsContinuationsIncludingEndedParents(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "ชื่อซ้ำ", Description: "ต้นทาง"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE researches SET status = ? WHERE id = ?`, "โครงการเสร็จสิ้น", parent.ID); err != nil {
		t.Fatalf("mark parent complete: %v", err)
	}

	for index := 0; index < 2; index++ {
		parentID := parent.ID
		child, err := repository.Create(context.Background(), repo.CreateResearchParams{
			Title:            "ชื่อซ้ำ",
			Description:      fmt.Sprintf("งานต่อเนื่อง %d", index+1),
			ContinuationOfID: &parentID,
		})
		if err != nil {
			t.Fatalf("create continuation %d: %v", index+1, err)
		}
		if child.ContinuationOfID == nil || *child.ContinuationOfID != parent.ID {
			t.Fatalf("continuationOfId = %v, want %d", child.ContinuationOfID, parent.ID)
		}
	}
}

func TestResearchRepositoryRejectsMissingContinuationWithoutPartialWrite(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	missingID := int64(999)

	_, err := repository.Create(context.Background(), repo.CreateResearchParams{
		Title:            "งานต่อเนื่อง",
		Description:      "รายละเอียด",
		ContinuationOfID: &missingID,
	})
	if !errors.Is(err, repo.ErrContinuationNotFound) {
		t.Fatalf("error = %v, want ErrContinuationNotFound", err)
	}
	assertResearchCount(t, db, 0)
}

func TestResearchRepositoryEnforcesRootAndContinuationTitleRules(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Parent", Description: "parent"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}

	_, err = repository.Create(context.Background(), repo.CreateResearchParams{Title: "Parent", Description: "duplicate root"})
	if !errors.Is(err, repo.ErrTitleAlreadyExists) {
		t.Fatalf("duplicate root error = %v, want ErrTitleAlreadyExists", err)
	}

	if _, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "parent", Description: "case-sensitive root"}); err != nil {
		t.Fatalf("case-sensitive title should be allowed: %v", err)
	}

	parentID := parent.ID
	if _, err := repository.Create(context.Background(), repo.CreateResearchParams{
		Title: "Reserved", Description: "continuation title", ContinuationOfID: &parentID,
	}); err != nil {
		t.Fatalf("create continuation with unique title: %v", err)
	}
	_, err = repository.Create(context.Background(), repo.CreateResearchParams{Title: "Reserved", Description: "root collides with continuation"})
	if !errors.Is(err, repo.ErrTitleAlreadyExists) {
		t.Fatalf("root colliding with continuation error = %v, want ErrTitleAlreadyExists", err)
	}
}

func TestResearchRepositoryNeverReusesIDsAndRejectsIdentityMutation(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	first, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "First", Description: "first"})
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `DELETE FROM researches WHERE id = ?`, first.ID); err != nil {
		t.Fatalf("delete first: %v", err)
	}
	second, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Second", Description: "second"})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}
	if second.ID <= first.ID {
		t.Fatalf("second id = %d, want greater than deleted id %d", second.ID, first.ID)
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE researches SET id = ? WHERE id = ?`, second.ID+100, second.ID); err == nil {
		t.Fatal("updating immutable id succeeded, want database error")
	}

	parent, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Parent", Description: "parent"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	parentID := parent.ID
	child, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Child", Description: "child", ContinuationOfID: &parentID})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE researches SET continuation_of_id = NULL WHERE id = ?`, child.ID); err == nil {
		t.Fatal("updating immutable continuation_of_id succeeded, want database error")
	}
}

func TestResearchDatabaseRejectsUnnormalizedOrInvalidStoredText(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
	}{
		{name: "title leading unicode whitespace", title: "\u00a0Title", description: "description"},
		{name: "description trailing unicode whitespace", title: "Title", description: "description\u3000"},
		{name: "empty title", title: "", description: "description"},
		{name: "title too long", title: repeatRune('ก', 201), description: "description"},
		{name: "description too long", title: "Title", description: repeatRune('ก', 5001)},
		{name: "title slash", title: "bad/title", description: "description"},
		{name: "title control", title: "bad\u0001title", description: "description"},
		{name: "description control", title: "Title", description: "bad\u0001description"},
		{name: "description carriage return", title: "Title", description: "bad\rdescription"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newResearchDatabase(t)
			_, err := db.ExecContext(context.Background(), `
				INSERT INTO researches (title, description, continuation_of_id)
				VALUES (?, ?, NULL)`, test.title, test.description)
			if err == nil {
				t.Fatalf("database accepted invalid title=%q description=%q", test.title, test.description)
			}
			assertResearchCount(t, db, 0)
		})
	}
}

func TestResearchDatabaseAcceptsNormalizedBoundariesAndAllowedDescriptionControls(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
	}{
		{name: "minimum", title: "ก", description: "ข"},
		{name: "maximum", title: repeatRune('ก', 200), description: repeatRune('ข', 5000)},
		{name: "description LF and tab", title: "Title", description: "line one\nline\ttwo"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newResearchDatabase(t)
			if _, err := db.ExecContext(context.Background(), `
				INSERT INTO researches (title, description, continuation_of_id)
				VALUES (?, ?, NULL)`, test.title, test.description); err != nil {
				t.Fatalf("database rejected valid title=%q description=%q: %v", test.title, test.description, err)
			}
			assertResearchCount(t, db, 1)
		})
	}
}

func TestResearchRepositoryConcurrentCreatesKeepIDsUnique(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	const workers = 12

	var wait sync.WaitGroup
	ids := make(chan int64, workers)
	errorsFound := make(chan error, workers)
	for index := 0; index < workers; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			created, err := repository.Create(context.Background(), repo.CreateResearchParams{
				Title: fmt.Sprintf("Research %02d", index), Description: "description",
			})
			if err != nil {
				errorsFound <- err
				return
			}
			ids <- created.ID
		}(index)
	}
	wait.Wait()
	close(ids)
	close(errorsFound)
	for err := range errorsFound {
		t.Errorf("concurrent create: %v", err)
	}
	seen := map[int64]bool{}
	for id := range ids {
		if id <= 0 || seen[id] {
			t.Errorf("generated id %d is non-positive or duplicated", id)
		}
		seen[id] = true
	}
	if len(seen) != workers {
		t.Fatalf("unique IDs = %d, want %d", len(seen), workers)
	}
}

func TestResearchRepositoryConcurrentDuplicateRootsAllowExactlyOne(t *testing.T) {
	db := newResearchDatabase(t)
	repository := repo.NewResearchRepository(db)
	start := make(chan struct{})
	results := make(chan error, 2)

	for index := 0; index < 2; index++ {
		go func() {
			<-start
			_, err := repository.Create(context.Background(), repo.CreateResearchParams{Title: "Same", Description: "description"})
			results <- err
		}()
	}
	close(start)

	successes := 0
	conflicts := 0
	for index := 0; index < 2; index++ {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, repo.ErrTitleAlreadyExists):
			conflicts++
		default:
			t.Fatalf("concurrent duplicate error = %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes/conflicts = %d/%d, want 1/1", successes, conflicts)
	}
	assertResearchCount(t, db, 1)
}

func newResearchDatabase(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "research.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	db.SetMaxOpenConns(12)
	if err := database.Initialize(context.Background(), db); err != nil {
		_ = db.Close()
		t.Fatalf("initialize schema: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertResearchCount(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM researches`).Scan(&count); err != nil {
		t.Fatalf("count researches: %v", err)
	}
	if count != want {
		t.Fatalf("research count = %d, want %d", count, want)
	}
}

func repeatRune(value rune, count int) string {
	result := make([]rune, count)
	for index := range result {
		result[index] = value
	}
	return string(result)
}
