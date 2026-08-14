package repo

import (
	"context"
	"database/sql"
)

type HealthRepository struct {
	database *sql.DB
}

func NewHealthRepository(database *sql.DB) *HealthRepository {
	return &HealthRepository{database: database}
}

func (repository *HealthRepository) Check(ctx context.Context) error {
	return repository.database.PingContext(ctx)
}
