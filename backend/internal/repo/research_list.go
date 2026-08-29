package repo

import (
	"context"
	"fmt"
)

func (repository *ResearchRepository) List(ctx context.Context) ([]Research, error) {
	rows, err := repository.database.QueryContext(ctx, `SELECT id FROM researches ORDER BY title ASC, id ASC`)
	if err != nil {
		return nil, fmt.Errorf("query researches: %w", err)
	}
	defer func() { _ = rows.Close() }()

	researches := make([]Research, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan research id: %w", err)
		}
		research, err := loadResearch(ctx, repository.database, id)
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
