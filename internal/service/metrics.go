package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/as-tanais/observy/internal/audit"
	models "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/repository"
)

// MetricsService управляет метриками: сохраняет, обновляет, возвращает.
// Поддерживает работу с разными типами хранилищ (память, файл, БД).
type MetricsService struct {
	storage       repository.Storage
	fileStorage   *repository.FileStorage
	storeInterval time.Duration
	observer      *audit.Observer
}

// NewMetricsService создаёт новый сервис метрик.
//
// Параметры:
//   - storage: основное хранилище метрик
//   - fileStorage: файловое хранилище для персистентности (опционально)
//   - storeInterval: интервал автосохранения в файл (0 = отключено)
//   - observer: наблюдатель для аудита операций (опционально)
//
// Возвращает указатель на инициализированный сервис.
func NewMetricsService(
	storage repository.Storage,
	fileStorage *repository.FileStorage,
	storeInterval time.Duration,
	observer *audit.Observer,
) *MetricsService {
	return &MetricsService{
		storage:       storage,
		fileStorage:   fileStorage,
		storeInterval: storeInterval,
		observer:      observer,
	}
}

// SetMetric устанавливает значение метрики.
//
// Поддерживаемые типы метрик:
//   - "counter": целочисленное значение, суммируется
//   - "gauge": дробное значение, перезаписывается
//
// Возвращает ошибку, если:
//   - имя метрики пустое
//   - неизвестный тип метрики
//   - ошибка сохранения в хранилище
func (s *MetricsService) SetMetric(ctx context.Context, metricType, metricName, valueStr, ipAddress string) error {
	if metricName == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	metric := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	var err error
	switch metricType {
	case models.Counter:
		err = s.setCounterMetric(ctx, &metric, valueStr)
	case models.Gauge:
		err = s.setGaugeMetric(ctx, &metric, valueStr)
	default:
		err = fmt.Errorf("unknown metric type: %s", metricType)
	}

	if err == nil {
		s.auditOperation(ctx, metricName, ipAddress)
	}

	return err
}

// GetMetric возвращает метрику по её имени из хранилища.
// Возвращает:
// - models.Metric
// - ошибку, если:
//   - имя метрики пустое
//   - метрики нет в БД
//   - ошибка получения из хранилища
//
// Пример:
//
//	metric, err := svc.GetMetric(ctx, "test_counter")
//	if err != nil {
//	log.Error("Failed to get metric", zap.Error(err))
//	return
//	}
//	if metric.MType == models.Counter && metric.Delta != nil {
//	fmt.Printf("Counter value: %d\n", *metric.Delta)
//	}
func (s *MetricsService) GetMetric(ctx context.Context, metricName string) (models.Metrics, error) {
	if metricName == "" {
		return models.Metrics{}, fmt.Errorf("metric name cannot be empty")
	}

	metric, err := s.storage.GetMetric(ctx, metricName)
	if err != nil {
		if errors.Is(err, repository.ErrMetricNotFound) {
			return models.Metrics{}, fmt.Errorf("metric '%s' not found", metricName)
		}
		return models.Metrics{}, fmt.Errorf("failed to get metric '%s': %w", metricName, err)
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
			return s.storage.SetMetric(ctx, metric)
		}
		return fmt.Errorf("failed to get existing counter: %w", err)
	}

	if existingMetric.Delta == nil {
		return fmt.Errorf("existing counter has nil delta")
	}
	newDelta := *existingMetric.Delta + v
	metric.Delta = &newDelta
	return s.storage.SetMetric(ctx, metric)
}

func (s *MetricsService) setGaugeMetric(ctx context.Context, metric *models.Metrics, valueStr string) error {
	v, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gauge value '%s': %w", valueStr, err)
	}
	metric.Value = &v
	return s.storage.SetMetric(ctx, metric)
}

// GetAllMetrics возращает все метрики которые есть в хранилище
func (s *MetricsService) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	return s.storage.GetAllMetrics(ctx)
}

func (s *MetricsService) SetNewMetric(ctx context.Context, input models.Metrics, ipAddress string) error {
	if input.ID == "" {
		return fmt.Errorf("metric name cannot be empty")
	}

	var err error
	switch input.MType {
	case models.Counter:
		if input.Delta == nil {
			return fmt.Errorf("delta value is required for counter metric")
		}

		existingMetric, getErr := s.storage.GetMetric(ctx, input.ID)
		if getErr != nil {
			if errors.Is(getErr, repository.ErrMetricNotFound) {
				err = s.storage.SetMetric(ctx, &input)
			} else {
				err = fmt.Errorf("failed to get existing counter: %w", getErr)
			}
		} else {
			if existingMetric.Delta == nil {
				return fmt.Errorf("existing counter has nil delta")
			}
			newDelta := *existingMetric.Delta + *input.Delta
			input.Delta = &newDelta
			err = s.storage.SetMetric(ctx, &input)
		}

	case models.Gauge:
		if input.Value == nil {
			return fmt.Errorf("value is required for gauge metric")
		}
		err = s.storage.SetMetric(ctx, &input)

	default:
		err = fmt.Errorf("unsupported metric type: %s", input.MType)
	}

	if err == nil {
		s.auditOperation(ctx, input.ID, ipAddress)
	}

	return err
}

// SaveToFile сохраняет все метрики в файл
// Возращает ошибку если:
//   - не удалось найти файл
//   - не удалось получить текущие метрики из хранилища
//   - не удалось сохранить метрики в файл
func (s *MetricsService) SaveToFile(ctx context.Context) error {
	metrics, err := s.storage.GetAllMetrics(ctx)
	if err != nil {
		return err
	}

	return s.fileStorage.SaveMetrics(metrics)
}

// LoadMetrics загружает метрики из файла
func (s *MetricsService) LoadMetrics(ctx context.Context) error {
	metrics, err := s.fileStorage.LoadMetrics()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		if err := s.storage.SetMetric(ctx, &m); err != nil {
			return err
		}
	}

	return nil
}

func (s *MetricsService) SetMetricWithSync(ctx context.Context, model models.Metrics, ipAddress string) error {
	if err := s.SetNewMetric(ctx, model, ipAddress); err != nil {
		return err
	}

	if s.storeInterval == 0 {
		return s.SaveToFile(ctx)
	}
	return nil
}

// UpdateBatch - сохранение набора метрик в хранилище
func (s *MetricsService) UpdateBatch(ctx context.Context, metrics []models.Metrics, ipAddress string) error {
	for _, m := range metrics {
		if err := s.SetNewMetric(ctx, m, ipAddress); err != nil {
			return err
		}
	}

	if s.observer != nil && ipAddress != "" {
		metricNames := make([]string, len(metrics))
		for i, m := range metrics {
			metricNames[i] = m.ID
		}
		s.observer.Notify(ctx, metricNames, ipAddress)
	}

	return nil
}

func (s *MetricsService) auditOperation(ctx context.Context, metricName string, ipAddress string) {
	if s.observer != nil && ipAddress != "" {
		s.observer.Notify(ctx, []string{metricName}, ipAddress)
	}
}
