package repository

import models "github.com/as-tanais/observy/internal/model"

type Storage interface {
	SetMetric(models.Metrics) error
	GetMetric(string) (models.Metrics, bool)
}
