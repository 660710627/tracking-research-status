package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrContinuationNotFound     = errors.New("continuation research not found")
	ErrResearchHasContinuations = errors.New("research has continuations")
	ErrResearchNotFound         = errors.New("research not found")
	ErrTitleAlreadyExists       = errors.New("research title already exists")
)

type Research struct {
	ID               int64
	Title            string
	Description      string
	ContinuationOfID *int64
	Status           string
	Process          string
}

type CreateResearchParams struct {
	Title            string
	Description      string
	ContinuationOfID *int64
}

type UpdateResearchParams struct {
	ID          int64
	Title       string
	Description string
}

type ResearchRepository struct {
	database   *sql.DB
	mutationMu sync.Mutex
}

func NewResearchRepository(database *sql.DB) *ResearchRepository {
	return &ResearchRepository{database: database}
}

func (repository *ResearchRepository) Create(ctx context.Context, params CreateResearchParams) (Research, error) {
	repository.mutationMu.Lock()
	defer repository.mutationMu.Unlock()

	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return Research{}, fmt.Errorf("begin create research: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	if params.ContinuationOfID != nil {
		var exists int
		err = transaction.QueryRowContext(ctx, `SELECT 1 FROM researches WHERE id = ?`, *params.ContinuationOfID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return Research{}, ErrContinuationNotFound
		}
		if err != nil {
			return Research{}, fmt.Errorf("check continuation research: %w", err)
		}
	}

	result, err := transaction.ExecContext(ctx, `
		INSERT INTO researches (title, description, continuation_of_id)
		VALUES (?, ?, ?)`, params.Title, params.Description, params.ContinuationOfID)
	if err != nil {
		return Research{}, mapCreateError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Research{}, fmt.Errorf("read created research id: %w", err)
	}

	created, err := scanResearch(transaction.QueryRowContext(ctx, `
		SELECT id, title, description, continuation_of_id, status, process
		FROM researches WHERE id = ?`, id))
	if err != nil {
		return Research{}, fmt.Errorf("read created research: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return Research{}, fmt.Errorf("commit created research: %w", err)
	}
	return created, nil
}

func (repository *ResearchRepository) Update(ctx context.Context, params UpdateResearchParams) (Research, error) {
	repository.mutationMu.Lock()
	defer repository.mutationMu.Unlock()

	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return Research{}, fmt.Errorf("begin update research: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	result, err := transaction.ExecContext(ctx, `
		UPDATE researches
		SET title = ?, description = ?
		WHERE id = ?`, params.Title, params.Description, params.ID)
	if err != nil {
		return Research{}, mapUpdateError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Research{}, fmt.Errorf("read updated research count: %w", err)
	}
	if rowsAffected == 0 {
		return Research{}, ErrResearchNotFound
	}

	updated, err := scanResearch(transaction.QueryRowContext(ctx, `
		SELECT id, title, description, continuation_of_id, status, process
		FROM researches WHERE id = ?`, params.ID))
	if err != nil {
		return Research{}, fmt.Errorf("read updated research: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return Research{}, fmt.Errorf("commit updated research: %w", err)
	}
	return updated, nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanResearch(row rowScanner) (Research, error) {
	var research Research
	var continuation sql.NullInt64
	if err := row.Scan(&research.ID, &research.Title, &research.Description, &continuation, &research.Status, &research.Process); err != nil {
		return Research{}, err
	}
	if continuation.Valid {
		value := continuation.Int64
		research.ContinuationOfID = &value
	}
	return research, nil
}

func mapCreateError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "CONTINUATION_NOT_FOUND"):
		return ErrContinuationNotFound
	case strings.Contains(message, "TITLE_ALREADY_EXISTS"):
		return ErrTitleAlreadyExists
	default:
		return fmt.Errorf("insert research: %w", err)
	}
}

func mapUpdateError(err error) error {
	if strings.Contains(err.Error(), "TITLE_ALREADY_EXISTS") {
		return ErrTitleAlreadyExists
	}
	return fmt.Errorf("update research: %w", err)
}
