package service

import (
	"errors"
	"strconv"

	models "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/repository"
)

type MetricsService struct {
	storage repository.Storage
}

func NewMetricsService(storage repository.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) SetMetric(metricType, metricName, valueStr string) error {
	var metric models.Metrics
	metric.ID = metricName
	metric.MType = metricType

	existingMetric, exists := s.storage.GetMetric(metricName)

	switch metricType {
	case models.Counter:
		v, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return errors.New("invalid counter value")
		}
		var newDelta int64
		if exists && existingMetric.Delta != nil {
			newDelta = *existingMetric.Delta + v
		} else {
			newDelta = v
		}
		metric.Delta = &newDelta
	case models.Gauge:
		v, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return errors.New("invalid gauge value")
		}
		metric.Value = &v
	default:
		return errors.New("unknown metric type")
	}

	return s.storage.SetMetric(metric)
}
