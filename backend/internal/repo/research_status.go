package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrProjectAlreadyEnded     = errors.New("project already ended")
)

type UpdateResearchStatusParams struct {
	ID     int64
	Status string
}

// UpdateStatus changes only status. SQLite triggers enforce the transition
// invariant as the final guard, while this transaction returns typed errors.
func (repository *ResearchRepository) UpdateStatus(ctx context.Context, params UpdateResearchStatusParams) (Research, error) {
	repository.mutationMu.Lock()
	defer repository.mutationMu.Unlock()

	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return Research{}, fmt.Errorf("begin update research status: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	current, err := loadResearch(ctx, transaction, params.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return Research{}, ErrResearchNotFound
	}
	if err != nil {
		return Research{}, fmt.Errorf("read current research status: %w", err)
	}
	if current.Status == params.Status {
		if err := transaction.Commit(); err != nil {
			return Research{}, fmt.Errorf("commit unchanged research status: %w", err)
		}
		return current, nil
	}

	if _, err := transaction.ExecContext(ctx, `UPDATE researches SET status = ? WHERE id = ?`, params.Status, params.ID); err != nil {
		return Research{}, mapStatusUpdateError(err)
	}
	updated, err := loadResearch(ctx, transaction, params.ID)
	if err != nil {
		return Research{}, fmt.Errorf("read updated research status: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return Research{}, fmt.Errorf("commit updated research status: %w", err)
	}
	return updated, nil
}

func mapStatusUpdateError(err error) error {
	switch {
	case strings.Contains(err.Error(), "PROJECT_ALREADY_ENDED"):
		return ErrProjectAlreadyEnded
	case strings.Contains(err.Error(), "INVALID_STATUS_TRANSITION"):
		return ErrInvalidStatusTransition
	default:
		return fmt.Errorf("update research status: %w", err)
	}
}
