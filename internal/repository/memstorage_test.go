package repository_test

import (
	"testing"

	models "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_SetMetric(t *testing.T) {
	tests := []struct {
		name    string
		m       models.Metrics
		wantErr bool
	}{
		{
			name: "успешное сохранение gauge метрики",
			m: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: floatPtr(123.45),
			},
			wantErr: false,
		},
		{
			name: "успешное сохранение counter метрики",
			m: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: int64Ptr(10),
			},
			wantErr: false,
		},
		{
			name: "метрика с пустым ID",
			m: models.Metrics{
				ID:    "",
				MType: models.Gauge,
				Value: floatPtr(100.0),
			},
			wantErr: false,
		},
		{
			name: "обновление существующей метрики",
			m: models.Metrics{
				ID:    "TestMetric",
				MType: models.Gauge,
				Value: floatPtr(200.0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage := repository.NewMemStorage()

			err := storage.SetMetric(tt.m)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем, что метрика сохранена
				retrieved, exists := storage.GetMetric(tt.m.ID)
				assert.True(t, exists, "Метрика должна существовать после SetMetric")
				assert.Equal(t, tt.m.ID, retrieved.ID)
				assert.Equal(t, tt.m.MType, retrieved.MType)
			}
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {
	tests := []struct {
		name       string
		setupData  []models.Metrics
		searchID   string
		wantExists bool
		wantMetric models.Metrics
	}{
		{
			name: "получение существующей gauge метрики",
			setupData: []models.Metrics{
				{
					ID:    "Alloc",
					MType: models.Gauge,
					Value: floatPtr(123.45),
				},
			},
			searchID:   "Alloc",
			wantExists: true,
			wantMetric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: floatPtr(123.45),
			},
		},
		{
			name: "получение существующей counter метрики",
			setupData: []models.Metrics{
				{
					ID:    "PollCount",
					MType: models.Counter,
					Delta: int64Ptr(42),
				},
			},
			searchID:   "PollCount",
			wantExists: true,
			wantMetric: models.Metrics{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: int64Ptr(42),
			},
		},
		{
			name:       "поиск несуществующей метрики",
			setupData:  []models.Metrics{},
			searchID:   "NonExistent",
			wantExists: false,
			wantMetric: models.Metrics{},
		},
		{
			name: "поиск метрики среди нескольких",
			setupData: []models.Metrics{
				{ID: "Metric1", MType: models.Gauge, Value: floatPtr(1.0)},
				{ID: "Metric2", MType: models.Gauge, Value: floatPtr(2.0)},
				{ID: "Metric3", MType: models.Gauge, Value: floatPtr(3.0)},
			},
			searchID:   "Metric2",
			wantExists: true,
			wantMetric: models.Metrics{
				ID:    "Metric2",
				MType: models.Gauge,
				Value: floatPtr(2.0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.NewMemStorage()

			for _, m := range tt.setupData {
				storage.SetMetric(m)
			}

			got, exists := storage.GetMetric(tt.searchID)

			assert.Equal(t, tt.wantExists, exists)

			if tt.wantExists {
				assert.Equal(t, tt.wantMetric.ID, got.ID)
				assert.Equal(t, tt.wantMetric.MType, got.MType)

				if tt.wantMetric.Value != nil {
					require.NotNil(t, got.Value)
					assert.Equal(t, *tt.wantMetric.Value, *got.Value)
				}

				if tt.wantMetric.Delta != nil {
					require.NotNil(t, got.Delta)
					assert.Equal(t, *tt.wantMetric.Delta, *got.Delta)
				}
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
