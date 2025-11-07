package repository

import (
	"context"
	"errors"

	models "github.com/as-tanais/observy/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/as-tanais/observy/pkg/helpers/retry"
)

type PGStorage struct {
	db *pgxpool.Pool
}

func NewPGStorage(db *pgxpool.Pool) Storage {
	return &PGStorage{db: db}
}

func (r *PGStorage) SetMetric(ctx context.Context, m models.Metrics) error {

	return retry.WithBackoff(func() error {
		_, err := r.db.Exec(ctx,
			`INSERT INTO metrics (id, m_type, delta, value, hash) 
             VALUES ($1, $2, $3, $4, $5)
             ON CONFLICT (id) DO UPDATE SET
                 m_type = $2,
                 delta = $3,
                 value = $4,
                 hash = $5`,
			m.ID, m.MType, m.Delta, m.Value, m.Hash,
		)
		return err
	})
}

func (r *PGStorage) GetMetric(ctx context.Context, id string) (models.Metrics, error) {

	var m models.Metrics

	err := retry.WithBackoff(func() error {
		return r.db.QueryRow(ctx,
			`SELECT id, m_type, delta, value, hash FROM metrics WHERE id = $1`,
			id,
		).Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Metrics{}, ErrMetricNotFound
		}
		return models.Metrics{}, err
	}
	return m, nil
}

func (r *PGStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	var metrics []models.Metrics

	err := retry.WithBackoff(func() error {
		rows, err := r.db.Query(ctx, `SELECT id, m_type, delta, value, hash FROM metrics`)
		if err != nil {
			return err
		}
		defer rows.Close()

		metrics, err = pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])
		return err
	})

	if err != nil {
		return nil, err
	}

	return metrics, nil
}
