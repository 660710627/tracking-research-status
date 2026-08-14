package service_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/660710627/my-research/internal/repo"
	"github.com/660710627/my-research/internal/service"
)

func TestResearchUpdateServiceNormalizesAndReplacesDetails(t *testing.T) {
	_ = newServiceTestDatabase(t)
	parentID := int64(7)
	var captured repo.UpdateResearchParams
	store := researchUpdateStoreStub{update: func(_ context.Context, params repo.UpdateResearchParams) (repo.Research, error) {
		captured = params
		return repo.Research{
			ID: params.ID, Title: params.Title, Description: params.Description, ContinuationOfID: &parentID,
			Status: "โครงการเสร็จสิ้น", Process: "บันทึกข้อตกลง",
		}, nil
	}}
	sut := service.NewResearchUpdateService(store)

	updated, err := sut.Update(context.Background(), service.UpdateResearchInput{
		ID: 42, Title: "\u00a0 ชื่องานวิจัย \u3000", Description: "\u2003บรรทัดแรก\r\nบรรทัดสอง\t\u2003",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if captured.ID != 42 || captured.Title != "ชื่องานวิจัย" || captured.Description != "บรรทัดแรก\nบรรทัดสอง" {
		t.Fatalf("normalized params = %#v", captured)
	}
	if updated.ID != 42 || updated.Title != captured.Title || updated.Description != captured.Description ||
		updated.ContinuationOfID == nil || *updated.ContinuationOfID != parentID ||
		updated.Status != "โครงการเสร็จสิ้น" || updated.Process != "บันทึกข้อตกลง" {
		t.Fatalf("updated = %#v, want complete repository result", updated)
	}
}

func TestResearchUpdateServiceAcceptsBoundariesAndAllowedCharacters(t *testing.T) {
	tests := []struct{ name, title, description string }{
		{name: "minimum", title: "ก", description: "ข"},
		{name: "maximum", title: strings.Repeat("ก", 200), description: strings.Repeat("ข", 5000)},
		{name: "description LF and tab", title: "Title", description: "line one\nline\ttwo"},
		{name: "case-sensitive title", title: "Research", description: "description"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			called := false
			store := researchUpdateStoreStub{update: func(_ context.Context, params repo.UpdateResearchParams) (repo.Research, error) {
				called = true
				return repo.Research{ID: params.ID, Title: params.Title, Description: params.Description}, nil
			}}
			_, err := service.NewResearchUpdateService(store).Update(context.Background(), service.UpdateResearchInput{ID: 1, Title: test.title, Description: test.description})
			if err != nil {
				t.Fatalf("valid input rejected: %v", err)
			}
			if !called {
				t.Fatal("repository was not called")
			}
		})
	}
}

func TestResearchUpdateServiceRejectsNonPositiveID(t *testing.T) {
	for _, id := range []int64{0, -1} {
		t.Run(strconv.FormatInt(id, 10), func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			called := false
			store := researchUpdateStoreStub{update: func(context.Context, repo.UpdateResearchParams) (repo.Research, error) {
				called = true
				return repo.Research{}, nil
			}}
			_, err := service.NewResearchUpdateService(store).Update(context.Background(), service.UpdateResearchInput{ID: id, Title: "Title", Description: "description"})
			if !errors.Is(err, service.ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
			if called {
				t.Fatal("repository called for invalid ID")
			}
		})
	}
}

func TestResearchUpdateServiceRejectsInvalidTitleAndDescription(t *testing.T) {
	tests := []struct{ name, title, description string }{
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
		{name: "description carriage return", title: "Title", description: "bad\rdescription"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			called := false
			store := researchUpdateStoreStub{update: func(context.Context, repo.UpdateResearchParams) (repo.Research, error) {
				called = true
				return repo.Research{}, nil
			}}
			_, err := service.NewResearchUpdateService(store).Update(context.Background(), service.UpdateResearchInput{ID: 1, Title: test.title, Description: test.description})
			if !errors.Is(err, service.ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
			if called {
				t.Fatal("repository called for invalid values")
			}
		})
	}
}

func TestResearchUpdateServiceMapsTypedRepositoryErrorsWithoutLeakingDetails(t *testing.T) {
	tests := []struct {
		name      string
		repoError error
		want      error
	}{
		{name: "research missing", repoError: repo.ErrResearchNotFound, want: service.ErrResearchNotFound},
		{name: "root title conflict", repoError: repo.ErrTitleAlreadyExists, want: service.ErrTitleAlreadyExists},
		{name: "database failure", repoError: errors.New("database secret: file path"), want: service.ErrInternal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			store := researchUpdateStoreStub{update: func(context.Context, repo.UpdateResearchParams) (repo.Research, error) {
				return repo.Research{}, test.repoError
			}}
			_, err := service.NewResearchUpdateService(store).Update(context.Background(), service.UpdateResearchInput{ID: 1, Title: "Title", Description: "description"})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if strings.Contains(err.Error(), "database secret") {
				t.Fatalf("internal detail leaked: %v", err)
			}
		})
	}
}

type researchUpdateStoreStub struct {
	update func(context.Context, repo.UpdateResearchParams) (repo.Research, error)
}

func (stub researchUpdateStoreStub) Update(ctx context.Context, params repo.UpdateResearchParams) (repo.Research, error) {
	return stub.update(ctx, params)
}
