package repo

import (
	"context"
	"fmt"
	"strings"
)

func (repository *ResearchRepository) Delete(ctx context.Context, id int64) error {
	repository.mutationMu.Lock()
	defer repository.mutationMu.Unlock()

	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete research: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	result, err := transaction.ExecContext(ctx, `DELETE FROM researches WHERE id = ?`, id)
	if err != nil {
		if strings.Contains(err.Error(), "RESEARCH_HAS_CONTINUATIONS") {
			return ErrResearchHasContinuations
		}
		return fmt.Errorf("delete research: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted research count: %w", err)
	}
	if rowsAffected == 0 {
		return ErrResearchNotFound
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit deleted research: %w", err)
	}
	return nil
}
