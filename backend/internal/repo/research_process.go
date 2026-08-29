package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidProcessTransition = errors.New("invalid process transition")

type UpdateResearchProcessParams struct {
	ID      int64
	Process string
}

func (repository *ResearchRepository) UpdateProcess(ctx context.Context, params UpdateResearchProcessParams) (Research, error) {
	repository.mutationMu.Lock()
	defer repository.mutationMu.Unlock()
	tx, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return Research{}, fmt.Errorf("begin update process: %w", err)
	}
	defer tx.Rollback()
	current, err := loadResearch(ctx, tx, params.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return Research{}, ErrResearchNotFound
	}
	if err != nil {
		return Research{}, fmt.Errorf("read process: %w", err)
	}
	if current.Process == params.Process {
		if err := tx.Commit(); err != nil {
			return Research{}, err
		}
		return current, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE researches SET process=? WHERE id=?`, params.Process, params.ID); err != nil {
		return Research{}, mapProcessUpdateError(err)
	}
	updated, err := loadResearch(ctx, tx, params.ID)
	if err != nil {
		return Research{}, err
	}
	if err := tx.Commit(); err != nil {
		return Research{}, err
	}
	return updated, nil
}

func mapProcessUpdateError(err error) error {
	if strings.Contains(err.Error(), "PROJECT_ALREADY_ENDED") {
		return ErrProjectAlreadyEnded
	}
	if strings.Contains(err.Error(), "INVALID_PROCESS_TRANSITION") {
		return ErrInvalidProcessTransition
	}
	return fmt.Errorf("update process: %w", err)
}
