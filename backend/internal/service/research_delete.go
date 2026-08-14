package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/660710627/my-research/internal/repo"
)

type ResearchDeleteStore interface {
	Delete(context.Context, int64) error
}

type ResearchDeleteService struct {
	store ResearchDeleteStore
}

func NewResearchDeleteService(store ResearchDeleteStore) *ResearchDeleteService {
	return &ResearchDeleteService{store: store}
}

func (service *ResearchDeleteService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrValidation
	}

	err := service.store.Delete(ctx, id)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repo.ErrResearchNotFound):
		return ErrResearchNotFound
	case errors.Is(err, repo.ErrResearchHasContinuations):
		return ErrResearchHasContinuations
	default:
		return fmt.Errorf("%w: delete research", ErrInternal)
	}
}
