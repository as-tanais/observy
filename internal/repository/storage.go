package repository

import (
	"context"

	models "github.com/as-tanais/observy/internal/model"
)

type Storage interface {
	SetMetric(ctx context.Context, m models.Metrics) error
	GetMetric(ctx context.Context, name string) (models.Metrics, error)
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)
}
