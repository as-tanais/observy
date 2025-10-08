package repository

import (
	models "github.com/as-tanais/observy/internal/model"
)

type MemStorage struct {
	metrics map[string]models.Metrics
}

func NewMemStorage() Storage {
	return &MemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (s *MemStorage) SetMetric(m models.Metrics) error {

	if s.metrics == nil {
		s.metrics = make(map[string]models.Metrics)
	}

	s.metrics[m.ID] = m

	return nil
}

func (s *MemStorage) GetMetric(id string) (models.Metrics, bool) {
	m, ok := s.metrics[id]
	return m, ok
}

func (s *MemStorage) GetAllMetrics() []models.Metrics {
	metrics := make([]models.Metrics, 0, len(s.metrics))

	for _, metric := range s.metrics {
		metrics = append(metrics, metric)
	}

	return metrics
}
