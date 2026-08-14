package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/660710627/my-research/internal/repo"
)

type UpdateResearchInput struct {
	ID          int64
	Title       string
	Description string
}

type ResearchUpdateStore interface {
	Update(context.Context, repo.UpdateResearchParams) (repo.Research, error)
}

type ResearchUpdateService struct {
	store ResearchUpdateStore
}

func NewResearchUpdateService(store ResearchUpdateStore) *ResearchUpdateService {
	return &ResearchUpdateService{store: store}
}

func (service *ResearchUpdateService) Update(ctx context.Context, input UpdateResearchInput) (Research, error) {
	title := strings.TrimFunc(input.Title, unicode.IsSpace)
	description := strings.ReplaceAll(input.Description, "\r\n", "\n")
	description = strings.TrimFunc(description, unicode.IsSpace)
	if input.ID <= 0 || !validResearchTitle(title) || !validResearchDescription(description) {
		return Research{}, ErrValidation
	}

	updated, err := service.store.Update(ctx, repo.UpdateResearchParams{
		ID: input.ID, Title: title, Description: description,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrResearchNotFound):
			return Research{}, ErrResearchNotFound
		case errors.Is(err, repo.ErrTitleAlreadyExists):
			return Research{}, ErrTitleAlreadyExists
		default:
			return Research{}, fmt.Errorf("%w: update research", ErrInternal)
		}
	}
	return Research{
		ID: updated.ID, Title: updated.Title, Description: updated.Description,
		ContinuationOfID: updated.ContinuationOfID, Status: updated.Status, Process: updated.Process,
	}, nil
}
