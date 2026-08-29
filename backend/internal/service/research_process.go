package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/660710627/my-research/internal/repo"
)

type UpdateResearchProcessInput struct {
	ID      int64
	Process string
}

type ResearchProcessStore interface {
	UpdateProcess(context.Context, repo.UpdateResearchProcessParams) (repo.Research, error)
}

type ResearchProcessService struct {
	store ResearchProcessStore
}

func NewResearchProcessService(store ResearchProcessStore) *ResearchProcessService {
	return &ResearchProcessService{store: store}
}

func (service *ResearchProcessService) UpdateProcess(ctx context.Context, input UpdateResearchProcessInput) (Research, error) {
	if input.ID <= 0 || !validResearchProcess(input.Process) {
		return Research{}, ErrValidation
	}
	updated, err := service.store.UpdateProcess(ctx, repo.UpdateResearchProcessParams{
		ID: input.ID, Process: input.Process,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrResearchNotFound):
			return Research{}, ErrResearchNotFound
		case errors.Is(err, repo.ErrInvalidProcessTransition):
			return Research{}, ErrInvalidProcessTransition
		case errors.Is(err, repo.ErrProjectAlreadyEnded):
			return Research{}, ErrProjectAlreadyEnded
		default:
			return Research{}, fmt.Errorf("%w: update research process", ErrInternal)
		}
	}
	return updated, nil
}

func validResearchProcess(value string) bool {
	for _, process := range []string{"สัญญาโครงการ", "บันทึกข้อตกลง", "เปิดบัญชีธนาคาร", "การเบิกจ่ายเงิน", "การจัดสรรค่าธรรมเนียม", "การติดตามส่งรายงาน", "รายงานสรุปการใช้เงิน", "การปิดบัญชีธนาคาร"} {
		if value == process {
			return true
		}
	}
	return false
}
