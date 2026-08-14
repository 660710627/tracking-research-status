package repo

import (
	"context"
	"fmt"
)

func (repository *ResearchRepository) List(ctx context.Context) ([]Research, error) {
	rows, err := repository.database.QueryContext(ctx, `
		SELECT id, title, description, continuation_of_id, status, process
		FROM researches
		ORDER BY title ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("query researches: %w", err)
	}
	defer func() { _ = rows.Close() }()

	researches := make([]Research, 0)
	for rows.Next() {
		research, err := scanResearch(rows)
		if err != nil {
			return nil, fmt.Errorf("scan research: %w", err)
		}
		researches = append(researches, research)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate researches: %w", err)
	}
	return researches, nil
}
