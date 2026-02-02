package repository

import (
	"context"

	models "github.com/as-tanais/observy/internal/model"
)

// Storage — контракт хранилища метрик
// // Реализации могут использовать разные бэкенды:
//   - память (MemStorage)
//   - файловую систему (FileStorage)
//   - PostgreSQL (PGStorage)
type Storage interface {
	// SetMetric сохраняет метрику.
	SetMetric(ctx context.Context, m *models.Metrics) error

	// GetMetric возвращает метрику по имени.
	GetMetric(ctx context.Context, name string) (models.Metrics, error)

	// GetAllMetrics возвращает все метрики.
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)
}
