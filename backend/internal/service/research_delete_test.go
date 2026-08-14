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

func TestResearchDeleteServiceDeletesPositiveID(t *testing.T) {
	_ = newServiceTestDatabase(t)
	var capturedID int64
	store := researchDeleteStoreStub{delete: func(_ context.Context, id int64) error {
		capturedID = id
		return nil
	}}

	if err := service.NewResearchDeleteService(store).Delete(context.Background(), 42); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if capturedID != 42 {
		t.Fatalf("repository ID = %d, want 42", capturedID)
	}
}

func TestResearchDeleteServiceRejectsNonPositiveID(t *testing.T) {
	for _, id := range []int64{0, -1} {
		t.Run(strconv.FormatInt(id, 10), func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			called := false
			store := researchDeleteStoreStub{delete: func(context.Context, int64) error {
				called = true
				return nil
			}}
			err := service.NewResearchDeleteService(store).Delete(context.Background(), id)
			if !errors.Is(err, service.ErrValidation) {
				t.Fatalf("error = %v, want ErrValidation", err)
			}
			if called {
				t.Fatal("repository called for invalid ID")
			}
		})
	}
}

func TestResearchDeleteServiceMapsTypedRepositoryErrorsWithoutLeakingDetails(t *testing.T) {
	tests := []struct {
		name      string
		repoError error
		want      error
	}{
		{name: "research missing", repoError: repo.ErrResearchNotFound, want: service.ErrResearchNotFound},
		{name: "research has continuations", repoError: repo.ErrResearchHasContinuations, want: service.ErrResearchHasContinuations},
		{name: "database failure", repoError: errors.New("database secret: file path"), want: service.ErrInternal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = newServiceTestDatabase(t)
			store := researchDeleteStoreStub{delete: func(context.Context, int64) error {
				return test.repoError
			}}
			err := service.NewResearchDeleteService(store).Delete(context.Background(), 1)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if strings.Contains(err.Error(), "database secret") {
				t.Fatalf("internal detail leaked: %v", err)
			}
		})
	}
}

type researchDeleteStoreStub struct {
	delete func(context.Context, int64) error
}

func (stub researchDeleteStoreStub) Delete(ctx context.Context, id int64) error {
	return stub.delete(ctx, id)
}
