package service_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
	_ "modernc.org/sqlite"
)

func TestResearchServiceNormalizesAndCreatesResearch(t *testing.T) {
	_ = newServiceTestDatabase(t)
	parentID := int64(42)
	var captured repo.CreateResearchParams
	store := researchStoreStub{create: func(_ context.Context, params repo.CreateResearchParams) (repo.Research, error) {
		captured = params
		return repo.Research{
			ID: 1, Title: params.Title, Description: params.Description,
			ContinuationOfID: params.ContinuationOfID,
			Status:           "กำลังดำเนินการ", Process: "สัญญาโครงการ",
		}, nil
	}}
	sut := service.NewResearchService(store)

	created, err := sut.Create(context.Background(), service.CreateResearchInput{
		Title: "\u00a0 ชื่องานวิจัย \u3000", Description: "\u2003บรรทัดแรก\r\nบรรทัดสอง\t\u2003", ContinuationOfID: &parentID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if captured.Title != "ชื่องานวิจัย" {
		t.Fatalf("normalized title = %q", captured.Title)
	}
	if captured.Description != "บรรทัดแรก\nบรรทัดสอง" {
		t.Fatalf("normalized description = %q, want LF-normalized and Unicode-trimmed text", captured.Description)
	}
	if captured.ContinuationOfID == nil || *captured.ContinuationOfID != parentID {
		t.Fatalf("captured continuation = %v, want %d", captured.ContinuationOfID, parentID)
	}
	if created.ID != 1 || created.Status != "กำลังดำเนินการ" || created.Process != "สัญญาโครงการ" {
		t.Fatalf("created = %#v, want repository result", created)
	}
}

func TestResearchServiceAcceptsTextBoundariesAndAllowedDescriptionCharacters(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
	}{
		{name: "minimum", title: "ก", description: "ข"},
		{name: "maximum", title: strings.Repeat("ก", 200), description: strings.Repeat("ข", 5000)},
		{name: "description LF and tab", title: "Title", description: "line one\nline\ttwo"},
		{name: "case-sensitive title", title: "Research", description: "description"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			called := false
			store := researchStoreStub{create: func(_ context.Context, params repo.CreateResearchParams) (repo.Research, error) {
				called = true
				return repo.Research{ID: 1, Title: params.Title, Description: params.Description, Status: "กำลังดำเนินการ", Process: "สัญญาโครงการ"}, nil
			}}
			sut := service.NewResearchService(store)
			if _, err := sut.Create(context.Background(), service.CreateResearchInput{Title: test.title, Description: test.description}); err != nil {
				t.Fatalf("valid input rejected: %v", err)
			}
			if !called {
				t.Fatal("repository was not called for valid input")
			}
		})
	}
}

func TestResearchServiceRejectsInvalidTitleAndDescription(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
	}{
		{name: "empty title", title: "", description: "description"},
		{name: "whitespace title", title: "\u00a0\u3000", description: "description"},
		{name: "title 201 runes", title: strings.Repeat("ก", 201), description: "description"},
		{name: "title slash", title: "bad/title", description: "description"},
		{name: "title newline", title: "bad\ntitle", description: "description"},
		{name: "title tab", title: "bad\ttitle", description: "description"},
		{name: "title NUL", title: "bad\x00title", description: "description"},
		{name: "title Unicode control", title: "bad\u0085title", description: "description"},
		{name: "empty description", title: "Title", description: ""},
		{name: "whitespace description", title: "Title", description: "\u00a0\u3000"},
		{name: "description 5001 runes", title: "Title", description: strings.Repeat("ก", 5001)},
		{name: "description NUL", title: "Title", description: "bad\x00description"},
		{name: "description control", title: "Title", description: "bad\u0001description"},
		{name: "description carriage return only", title: "Title", description: "bad\rdescription"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			called := false
			store := researchStoreStub{create: func(_ context.Context, _ repo.CreateResearchParams) (repo.Research, error) {
				called = true
				return repo.Research{}, nil
			}}
			sut := service.NewResearchService(store)
			_, err := sut.Create(context.Background(), service.CreateResearchInput{Title: test.title, Description: test.description})
			if !errors.Is(err, service.ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
			if called {
				t.Fatal("repository called for invalid input")
			}
		})
	}
}

func TestResearchServiceRejectsNonPositiveContinuationID(t *testing.T) {
	for _, value := range []int64{0, -1} {
		t.Run(strconv.FormatInt(value, 10), func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			called := false
			store := researchStoreStub{create: func(_ context.Context, _ repo.CreateResearchParams) (repo.Research, error) {
				called = true
				return repo.Research{}, nil
			}}
			sut := service.NewResearchService(store)
			_, err := sut.Create(context.Background(), service.CreateResearchInput{Title: "Title", Description: "description", ContinuationOfID: &value})
			if !errors.Is(err, service.ErrValidation) {
				t.Fatalf("continuationOfId %d error = %v, want ErrValidation", value, err)
			}
			if called {
				t.Fatal("repository called for invalid continuationOfId")
			}
		})
	}
}

func TestResearchServiceMapsTypedRepositoryErrors(t *testing.T) {
	tests := []struct {
		name      string
		repoError error
		want      error
	}{
		{name: "continuation missing", repoError: repo.ErrContinuationNotFound, want: service.ErrContinuationNotFound},
		{name: "root title conflict", repoError: repo.ErrTitleAlreadyExists, want: service.ErrTitleAlreadyExists},
		{name: "unexpected database failure", repoError: errors.New("database secret"), want: service.ErrInternal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			store := researchStoreStub{create: func(_ context.Context, _ repo.CreateResearchParams) (repo.Research, error) {
				return repo.Research{}, test.repoError
			}}
			sut := service.NewResearchService(store)
			_, err := sut.Create(context.Background(), service.CreateResearchInput{Title: "Title", Description: "description"})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

type researchStoreStub struct {
	create func(context.Context, repo.CreateResearchParams) (repo.Research, error)
}

func (stub researchStoreStub) Create(ctx context.Context, params repo.CreateResearchParams) (repo.Research, error) {
	return stub.create(ctx, params)
}

func newServiceTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "service.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Fatalf("ping SQLite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
