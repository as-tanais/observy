package repository

import (
	"context"
	"sync"

	models "github.com/as-tanais/observy/internal/model"
)

type MemStorage struct {
	mu      sync.RWMutex
	metrics map[string]models.Metrics
}

func NewMemStorage() Storage {
	return &MemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (s *MemStorage) SetMetric(_ context.Context, m models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics[m.ID] = m

	return nil
}

func (s *MemStorage) GetMetric(_ context.Context, id string) (models.Metrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.metrics[id]

	if !ok {
		return models.Metrics{}, ErrMetricNotFound
	}

	return m, nil
}

func (s *MemStorage) GetAllMetrics(_ context.Context) ([]models.Metrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(s.metrics))
	for _, metric := range s.metrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
