package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	models "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/repository"
)

type MetricsService struct {
	storage       repository.Storage
	fileStorage   *repository.FileStorage
	storeInterval time.Duration
}

func NewMetricsService(storage repository.Storage, fileStorage *repository.FileStorage, storeInterval time.Duration) *MetricsService {
	return &MetricsService{
		storage:       storage,
		fileStorage:   fileStorage,
		storeInterval: storeInterval,
	}
}

func (s *MetricsService) SetMetric(ctx context.Context, metricType, metricName, valueStr string) error {
	if metricName == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	metric := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case models.Counter:
		return s.setCounterMetric(ctx, &metric, valueStr)
	case models.Gauge:
		return s.setGaugeMetric(ctx, &metric, valueStr)
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
}

func (s *MetricsService) GetMetric(ctx context.Context, metricName string) (models.Metrics, error) {
	if metricName == "" {
		return models.Metrics{}, fmt.Errorf("metric name cannot be empty")
	}

	metric, err := s.storage.GetMetric(ctx, metricName)
	if err != nil {
		return models.Metrics{}, fmt.Errorf("metric '%s' not found", metricName)
	}

	return metric, nil
}

func (s *MetricsService) setCounterMetric(ctx context.Context, metric *models.Metrics, valueStr string) error {
	v, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value '%s': %w", valueStr, err)
	}

	existingMetric, err := s.storage.GetMetric(ctx, metric.ID)
	if err != nil {
		if errors.Is(err, repository.ErrMetricNotFound) {
			newDelta := v
			metric.Delta = &newDelta
			return s.storage.SetMetric(ctx, *metric)
		}
		return fmt.Errorf("failed to get existing counter: %w", err)
	}

	if existingMetric.Delta == nil {
		return fmt.Errorf("existing counter has nil delta")
	}
	newDelta := *existingMetric.Delta + v
	metric.Delta = &newDelta
	return s.storage.SetMetric(ctx, *metric)
}

func (s *MetricsService) setGaugeMetric(ctx context.Context, metric *models.Metrics, valueStr string) error {
	v, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gauge value '%s': %w", valueStr, err)
	}
	metric.Value = &v
	return s.storage.SetMetric(ctx, *metric)
}

func (s *MetricsService) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	return s.storage.GetAllMetrics(ctx)
}

func (s *MetricsService) SetNewMetric(ctx context.Context, input models.Metrics) error {
	if input.ID == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	switch input.MType {
	case models.Counter:
		if input.Delta == nil {
			return fmt.Errorf("delta value is required for counter metric")
		}

		existingMetric, err := s.storage.GetMetric(ctx, input.ID)
		if err != nil {
			if errors.Is(err, repository.ErrMetricNotFound) {
				return s.storage.SetMetric(ctx, input)
			}
			return fmt.Errorf("failed to get existing counter: %w", err)
		}

		if existingMetric.Delta == nil {
			return fmt.Errorf("existing counter has nil delta")
		}
		newDelta := *existingMetric.Delta + *input.Delta
		input.Delta = &newDelta
		return s.storage.SetMetric(ctx, input)

	case models.Gauge:
		if input.Value == nil {
			return fmt.Errorf("value is required for gauge metric")
		}
		return s.storage.SetMetric(ctx, input)

	default:
		return fmt.Errorf("unsupported metric type: %s", input.MType)
	}
}

func (s *MetricsService) SaveToFile(ctx context.Context) error {
	metrics, err := s.storage.GetAllMetrics(ctx)
	if err != nil {
		return err
	}

	return s.fileStorage.SaveMetrics(metrics)
}

func (s *MetricsService) LoadMetrics(ctx context.Context) error {
	metrics, err := s.fileStorage.LoadMetrics()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		if err := s.storage.SetMetric(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

func (s *MetricsService) SetMetricWithSync(ctx context.Context, model models.Metrics) error {
	if err := s.SetNewMetric(ctx, model); err != nil {
		return err
	}

	if s.storeInterval == 0 {
		return s.SaveToFile(ctx)
	}
	return nil
}
