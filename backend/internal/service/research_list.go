package service

import (
	"context"
	"fmt"

	"github.com/660710627/my-research/internal/repo"
)

type ResearchListStore interface {
	List(context.Context) ([]repo.Research, error)
}

type ResearchListService struct {
	store ResearchListStore
}

func NewResearchListService(store ResearchListStore) *ResearchListService {
	return &ResearchListService{store: store}
}

func (service *ResearchListService) List(ctx context.Context) ([]Research, error) {
	stored, err := service.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: list researches", ErrInternal)
	}
	return stored, nil
}
