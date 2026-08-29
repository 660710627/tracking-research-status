package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/660710627/my-research/internal/repo"
)

type UpdateResearchStatusInput struct {
	ID     int64
	Status string
}

type ResearchStatusStore interface {
	UpdateStatus(context.Context, repo.UpdateResearchStatusParams) (repo.Research, error)
}

type ResearchStatusService struct {
	store ResearchStatusStore
}

func NewResearchStatusService(store ResearchStatusStore) *ResearchStatusService {
	return &ResearchStatusService{store: store}
}

func (service *ResearchStatusService) UpdateStatus(ctx context.Context, input UpdateResearchStatusInput) (Research, error) {
	if input.ID <= 0 || !validResearchStatus(input.Status) {
		return Research{}, ErrValidation
	}
	updated, err := service.store.UpdateStatus(ctx, repo.UpdateResearchStatusParams{
		ID: input.ID, Status: input.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrResearchNotFound):
			return Research{}, ErrResearchNotFound
		case errors.Is(err, repo.ErrInvalidStatusTransition):
			return Research{}, ErrInvalidStatusTransition
		case errors.Is(err, repo.ErrProjectAlreadyEnded):
			return Research{}, ErrProjectAlreadyEnded
		default:
			return Research{}, fmt.Errorf("%w: update research status", ErrInternal)
		}
	}
	return updated, nil
}

func validResearchStatus(status string) bool {
	switch status {
	case "กำลังดำเนินการ",
		"กำลังดำเนินการ (ขยายเวลาครั้งที่ 1)",
		"กำลังดำเนินการ (ขยายเวลาครั้งที่ 2)",
		"กำลังดำเนินการ (ขยายเวลามากกว่า 2 ครั้ง)",
		"โครงการเสร็จสิ้น",
		"ยุติโครงการ":
		return true
	default:
		return false
	}
}
