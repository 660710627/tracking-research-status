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

	researches := make([]Research, 0, len(stored))
	for _, research := range stored {
		researches = append(researches, Research{
			ID: research.ID, Title: research.Title, Description: research.Description,
			ContinuationOfID: research.ContinuationOfID, Status: research.Status, Process: research.Process,
		})
	}
	return researches, nil
}
