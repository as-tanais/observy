package repository

import (
	"context"

	models "github.com/as-tanais/observy/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStorage struct {
	db *pgxpool.Pool
}

func NewPGStorage(db *pgxpool.Pool) Storage {
	return &PGStorage{
		db: db,
	}
}

func (r *PGStorage) SetMetric(ctx context.Context, m models.Metrics) error {
	return nil
}
func (r *PGStorage) GetMetric(ctx context.Context, name string) (models.Metrics, error) {
	return models.Metrics{}, nil
}
func (r *PGStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	return nil, nil
}
