package service

import (
	"fmt"
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
	if metricName == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	metric := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case models.Counter:
		return s.setCounterMetric(&metric, valueStr)
	case models.Gauge:
		return s.setGaugeMetric(&metric, valueStr)
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
}

func (s *MetricsService) GetMetric(metricName string) (models.Metrics, error) {
	if metricName == "" {
		return models.Metrics{}, fmt.Errorf("metric name cannot be empty")
	}

	metric, exists := s.storage.GetMetric(metricName)
	if !exists {
		return models.Metrics{}, fmt.Errorf("metric '%s' not found", metricName)
	}

	return metric, nil
}

func (s *MetricsService) setCounterMetric(metric *models.Metrics, valueStr string) error {
	v, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value '%s': %w", valueStr, err)
	}

	existingMetric, exists := s.storage.GetMetric(metric.ID)

	var newDelta int64
	if exists && existingMetric.Delta != nil {
		newDelta = *existingMetric.Delta + v
	} else {
		newDelta = v
	}

	metric.Delta = &newDelta
	return s.storage.SetMetric(*metric)
}

func (s *MetricsService) setGaugeMetric(metric *models.Metrics, valueStr string) error {
	v, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gauge value '%s': %w", valueStr, err)
	}

	metric.Value = &v
	return s.storage.SetMetric(*metric)
}

func (s *MetricsService) GetAllMetrics() []models.Metrics {
	return s.storage.GetAllMetrics()
}

func (s *MetricsService) SetNewMetric(input models.Metrics) error {
	if input.ID == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	switch input.MType {
	case "counter":

		if input.Delta == nil {
			return fmt.Errorf("delta value is required for counter metric")
		}

		existingMetric, exists := s.storage.GetMetric(input.ID)

		var newDelta int64
		if exists && existingMetric.Delta != nil {

			newDelta = *existingMetric.Delta + *input.Delta
		} else {

			newDelta = *input.Delta
		}

		input.Delta = &newDelta
		return s.storage.SetMetric(input)

	case "gauge":

		if input.Value == nil {
			return fmt.Errorf("value is required for gauge metric")
		}

		return s.storage.SetMetric(input)

	default:
		return fmt.Errorf("unsupported metric type: %s", input.MType)
	}
}
