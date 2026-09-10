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

func (r *HealthRepository) Ping(ctx context.Context) error {
	return r.database.PingContext(ctx)
}
