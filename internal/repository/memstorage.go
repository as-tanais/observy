package repository

import models "github.com/as-tanais/observy/internal/model"

type MemStorage struct {
	metrics map[string]models.Metrics
}

type Storage interface {
	SetMetric(models.Metrics) error
	GetMetric(string) (models.Metrics, bool)
}

func NewStorage() *MemStorage {
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
