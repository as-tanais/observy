package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/repository"
)

func setupPGStorage(t *testing.T) (*pgxpool.Pool, func()) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	err = pool.Ping(ctx)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

func cleanupMetrics(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	_, err := pool.Exec(ctx, "TRUNCATE TABLE metrics")
	require.NoError(t, err)
}

func TestPGStorage_SetAndGetGauge(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	metric := models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: ptrFloat64(123.45),
	}

	err := storage.SetMetric(ctx, &metric)
	require.NoError(t, err)

	result, err := storage.GetMetric(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, "Alloc", result.ID)
	assert.Equal(t, models.Gauge, result.MType)
	require.NotNil(t, result.Value)
	assert.Equal(t, 123.45, *result.Value)
}

func TestPGStorage_SetAndGetCounter(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	metric := models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: ptrInt64(10),
	}

	err := storage.SetMetric(ctx, &metric)
	require.NoError(t, err)

	result, err := storage.GetMetric(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, "PollCount", result.ID)
	assert.Equal(t, models.Counter, result.MType)
	require.NotNil(t, result.Delta)
	assert.Equal(t, int64(10), *result.Delta)
}

func TestPGStorage_GetNonExistent(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	_, err := storage.GetMetric(ctx, "NonExistent")
	require.Error(t, err)

	assert.True(t, errors.Is(err, repository.ErrMetricNotFound))
}

func TestPGStorage_CounterAccumulation(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	metric1 := models.Metrics{
		ID:    "Counter",
		MType: models.Counter,
		Delta: ptrInt64(5),
	}
	err := storage.SetMetric(ctx, &metric1)
	require.NoError(t, err)

	metric2 := models.Metrics{
		ID:    "Counter",
		MType: models.Counter,
		Delta: ptrInt64(3),
	}
	err = storage.SetMetric(ctx, &metric2)
	require.NoError(t, err)

	result, err := storage.GetMetric(ctx, "Counter")
	require.NoError(t, err)
	require.NotNil(t, result.Delta)

	assert.Equal(t, int64(8), *result.Delta)
}

func TestPGStorage_GaugeOverwrite(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	// Первая запись
	metric1 := models.Metrics{
		ID:    "Gauge",
		MType: models.Gauge,
		Value: ptrFloat64(100),
	}
	err := storage.SetMetric(ctx, &metric1)
	require.NoError(t, err)

	metric2 := models.Metrics{
		ID:    "Gauge",
		MType: models.Gauge,
		Value: ptrFloat64(200),
	}
	err = storage.SetMetric(ctx, &metric2)
	require.NoError(t, err)

	result, err := storage.GetMetric(ctx, "Gauge")
	require.NoError(t, err)
	require.NotNil(t, result.Value)

	assert.Equal(t, 200.0, *result.Value)
}

func TestPGStorage_GetAll(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	metricsToAdd := []models.Metrics{
		{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: ptrFloat64(100.5),
		},
		{
			ID:    "TotalAlloc",
			MType: models.Gauge,
			Value: ptrFloat64(200.3),
		},
		{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: ptrInt64(42),
		},
	}

	for _, m := range metricsToAdd {
		err := storage.SetMetric(ctx, &m)
		require.NoError(t, err)
	}

	allMetrics, err := storage.GetAllMetrics(ctx)
	require.NoError(t, err)
	assert.Len(t, allMetrics, len(metricsToAdd))

	for _, expected := range metricsToAdd {
		found := false
		for _, actual := range allMetrics {
			if actual.ID == expected.ID {
				assert.Equal(t, expected.MType, actual.MType)
				found = true
				break
			}
		}
		assert.True(t, found, "Метрика %s не найдена в GetAll", expected.ID)
	}
}

func TestPGStorage_GetAllEmpty(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	allMetrics, err := storage.GetAllMetrics(ctx)
	require.NoError(t, err)
	assert.Len(t, allMetrics, 0)
}

func TestPGStorage_WithHash(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	metric := models.Metrics{
		ID:    "WithHash",
		MType: models.Gauge,
		Value: ptrFloat64(123.45),
		Hash:  "somehash123",
	}

	err := storage.SetMetric(ctx, &metric)
	require.NoError(t, err)

	result, err := storage.GetMetric(ctx, "WithHash")
	require.NoError(t, err)
	assert.Equal(t, "somehash123", result.Hash)
}

func TestPGStorage_MultipleMetricsUpdate(t *testing.T) {
	pool, cleanup := setupPGStorage(t)
	defer cleanup()
	cleanupMetrics(t, pool)

	ctx := context.Background()
	storage := repository.NewPGStorage(pool)

	metrics := []models.Metrics{
		{ID: "Metric1", MType: models.Gauge, Value: ptrFloat64(1.0)},
		{ID: "Metric2", MType: models.Gauge, Value: ptrFloat64(2.0)},
		{ID: "Metric3", MType: models.Counter, Delta: ptrInt64(3)},
	}

	for _, m := range metrics {
		err := storage.SetMetric(ctx, &m)
		require.NoError(t, err)
	}

	updated := models.Metrics{
		ID:    "Metric1",
		MType: models.Gauge,
		Value: ptrFloat64(10.0),
	}
	err := storage.SetMetric(ctx, &updated)
	require.NoError(t, err)

	m1, err := storage.GetMetric(ctx, "Metric1")
	require.NoError(t, err)
	assert.Equal(t, 10.0, *m1.Value)

	m2, err := storage.GetMetric(ctx, "Metric2")
	require.NoError(t, err)
	assert.Equal(t, 2.0, *m2.Value)

	m3, err := storage.GetMetric(ctx, "Metric3")
	require.NoError(t, err)
	assert.Equal(t, int64(3), *m3.Delta)
}

// === Helpers ===

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
