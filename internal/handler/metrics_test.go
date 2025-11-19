package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/as-tanais/observy/internal/handler"
	model "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
)

func newTestService(t *testing.T) *service.MetricsService {
	storage := repository.NewMemStorage()

	tmpFile, err := os.CreateTemp("", "metrics-test-*.json")
	require.NoError(t, err)
	defer tmpFile.Close()

	fileStorage := repository.NewFileStorage(tmpFile.Name())

	svc := service.NewMetricsService(storage, fileStorage, 100*time.Second)

	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return svc
}

func setupRouter(h *handler.MetricsHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Get("/value/{type}/{name}", h.GetMetricHandler)
	r.Post("/update", h.UpdateHandler)
	r.Post("/value", h.GetMetric)
	r.Post("/updates", h.UpdateMetricsHandler)
	r.Get("/", h.ListMetricsHandler)
	return r
}

// === Тесты для URL-based endpoints ===

func TestMetricsHandler_UpdateMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "успешное обновление gauge метрики",
			method:         http.MethodPost,
			path:           "/update/gauge/Alloc/123.45",
			wantStatusCode: http.StatusOK,
			wantBody:       "Metric updated",
		},
		{
			name:           "успешное обновление counter метрики",
			method:         http.MethodPost,
			path:           "/update/counter/PollCount/10",
			wantStatusCode: http.StatusOK,
			wantBody:       "Metric updated",
		},
		{
			name:           "неправильный HTTP метод (GET вместо POST)",
			method:         http.MethodGet,
			path:           "/update/gauge/Alloc/123.45",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "пустой тип метрики",
			method:         http.MethodPost,
			path:           "/update//Alloc/123",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "пустое имя метрики",
			method:         http.MethodPost,
			path:           "/update/gauge//123",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			h := handler.NewMetricsHandler(svc)
			router := setupRouter(h)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestMetricsHandler_GetMetricHandler(t *testing.T) {
	tests := []struct {
		name         string
		setupMetrics []struct {
			metricType string
			name       string
			value      string
		}
		method         string
		path           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "успешное получение gauge метрики",
			setupMetrics: []struct {
				metricType string
				name       string
				value      string
			}{
				{metricType: model.Gauge, name: "Alloc", value: "123.45"},
			},
			method:         http.MethodGet,
			path:           "/value/gauge/Alloc",
			wantStatusCode: http.StatusOK,
			wantBody:       "123.45",
		},
		{
			name: "успешное получение counter метрики",
			setupMetrics: []struct {
				metricType string
				name       string
				value      string
			}{
				{metricType: model.Counter, name: "PollCount", value: "42"},
			},
			method:         http.MethodGet,
			path:           "/value/counter/PollCount",
			wantStatusCode: http.StatusOK,
			wantBody:       "42",
		},
		{
			name:           "метрика не найдена",
			setupMetrics:   []struct{ metricType, name, value string }{},
			method:         http.MethodGet,
			path:           "/value/gauge/NonExistent",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "несоответствие типа метрики",
			setupMetrics: []struct {
				metricType string
				name       string
				value      string
			}{
				{metricType: model.Counter, name: "PollCount", value: "10"},
			},
			method:         http.MethodGet,
			path:           "/value/gauge/PollCount",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "неправильный HTTP метод",
			setupMetrics:   []struct{ metricType, name, value string }{},
			method:         http.MethodPost,
			path:           "/value/gauge/Alloc",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			h := handler.NewMetricsHandler(svc)

			// Предзагрузка метрик
			for _, m := range tt.setupMetrics {
				err := svc.SetMetric(context.Background(), m.metricType, m.name, m.value)
				require.NoError(t, err)
			}

			router := setupRouter(h)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// === Тесты для JSON-based endpoints ===

func TestMetricsHandler_UpdateHandler(t *testing.T) {
	tests := []struct {
		name           string
		input          model.Metrics
		wantStatusCode int
	}{
		{
			name: "успешное обновление gauge через JSON",
			input: model.Metrics{
				ID:    "Alloc",
				MType: model.Gauge,
				Value: ptrFloat64(123.45),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "успешное обновление counter через JSON",
			input: model.Metrics{
				ID:    "PollCount",
				MType: model.Counter,
				Delta: ptrInt64(10),
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			h := handler.NewMetricsHandler(svc)
			router := setupRouter(h)

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}

func TestMetricsHandler_GetMetric(t *testing.T) {
	tests := []struct {
		name         string
		setupMetrics []struct {
			id    string
			mType string
			value float64
			delta int64
		}
		input          model.Metrics
		wantStatusCode int
		wantMetric     *model.Metrics
	}{
		{
			name: "успешное получение gauge метрики через JSON",
			setupMetrics: []struct {
				id    string
				mType string
				value float64
				delta int64
			}{
				{id: "Alloc", mType: model.Gauge, value: 123.45},
			},
			input: model.Metrics{
				ID:    "Alloc",
				MType: model.Gauge,
			},
			wantStatusCode: http.StatusOK,
			wantMetric: &model.Metrics{
				ID:    "Alloc",
				MType: model.Gauge,
				Value: ptrFloat64(123.45),
			},
		},
		{
			name: "успешное получение counter метрики через JSON",
			setupMetrics: []struct {
				id    string
				mType string
				value float64
				delta int64
			}{
				{id: "PollCount", mType: model.Counter, delta: 42},
			},
			input: model.Metrics{
				ID:    "PollCount",
				MType: model.Counter,
			},
			wantStatusCode: http.StatusOK,
			wantMetric: &model.Metrics{
				ID:    "PollCount",
				MType: model.Counter,
				Delta: ptrInt64(42),
			},
		},
		{
			name: "метрика не найдена",
			setupMetrics: []struct {
				id, mType string
				value     float64
				delta     int64
			}{},
			input:          model.Metrics{ID: "NonExistent"},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "несоответствие типа метрики",
			setupMetrics: []struct {
				id    string
				mType string
				value float64
				delta int64
			}{
				{id: "PollCount", mType: model.Counter, delta: 10},
			},
			input: model.Metrics{
				ID:    "PollCount",
				MType: model.Gauge,
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "пустое имя метрики",
			setupMetrics: []struct {
				id, mType string
				value     float64
				delta     int64
			}{},
			input:          model.Metrics{ID: ""},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			h := handler.NewMetricsHandler(svc)

			// Предзагрузка метрик
			for _, m := range tt.setupMetrics {
				var value string
				if m.mType == model.Gauge {
					value = fmt.Sprintf("%g", m.value)
				} else {
					value = fmt.Sprintf("%d", m.delta)
				}
				err := svc.SetMetric(context.Background(), m.mType, m.id, value)
				require.NoError(t, err)
			}

			router := setupRouter(h)

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			if tt.wantMetric != nil && tt.wantStatusCode == http.StatusOK {
				var result model.Metrics
				err := json.NewDecoder(w.Body).Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.wantMetric.ID, result.ID)
				assert.Equal(t, tt.wantMetric.MType, result.MType)
				if tt.wantMetric.Value != nil {
					assert.NotNil(t, result.Value)
					assert.Equal(t, *tt.wantMetric.Value, *result.Value)
				}
				if tt.wantMetric.Delta != nil {
					assert.NotNil(t, result.Delta)
					assert.Equal(t, *tt.wantMetric.Delta, *result.Delta)
				}
			}
		})
	}
}

func TestMetricsHandler_UpdateMetricsHandler(t *testing.T) {
	tests := []struct {
		name           string
		input          []model.Metrics
		wantStatusCode int
	}{
		{
			name: "успешное обновление batch",
			input: []model.Metrics{
				{
					ID:    "Alloc",
					MType: model.Gauge,
					Value: ptrFloat64(123.45),
				},
				{
					ID:    "PollCount",
					MType: model.Counter,
					Delta: ptrInt64(10),
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "пустой batch",
			input:          []model.Metrics{},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			h := handler.NewMetricsHandler(svc)
			router := setupRouter(h)

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}

func TestMetricsHandler_ListMetricsHandler(t *testing.T) {
	tests := []struct {
		name         string
		setupMetrics []struct {
			id    string
			mType string
			value string
		}
		wantStatusCode int
	}{
		{
			name: "получение списка метрик",
			setupMetrics: []struct {
				id    string
				mType string
				value string
			}{
				{id: "Alloc", mType: model.Gauge, value: "123.45"},
				{id: "PollCount", mType: model.Counter, value: "42"},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "получение пустого списка",
			setupMetrics:   []struct{ id, mType, value string }{},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			h := handler.NewMetricsHandler(svc)

			// Предзагрузка метрик
			for _, m := range tt.setupMetrics {
				err := svc.SetMetric(context.Background(), m.mType, m.id, m.value)
				require.NoError(t, err)
			}

			router := setupRouter(h)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

			if len(tt.setupMetrics) > 0 {
				assert.Contains(t, w.Body.String(), "Metrics")
				for _, m := range tt.setupMetrics {
					assert.Contains(t, w.Body.String(), m.id)
					assert.Contains(t, w.Body.String(), m.mType)
				}
			} else {
				assert.Contains(t, w.Body.String(), "Метрики еще не подгружались")
			}
		})
	}
}

// === Интеграционные тесты ===

func TestMetricsHandler_UpdateAndGet(t *testing.T) {
	svc := newTestService(t)
	h := handler.NewMetricsHandler(svc)
	router := setupRouter(h)

	// Update через URL
	reqUpdate := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/999.99", nil)
	wUpdate := httptest.NewRecorder()
	router.ServeHTTP(wUpdate, reqUpdate)
	assert.Equal(t, http.StatusOK, wUpdate.Code)

	// Get через URL
	reqGet := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
	assert.Contains(t, wGet.Body.String(), "999.99")
}

func TestMetricsHandler_CounterAccumulation(t *testing.T) {
	svc := newTestService(t)
	h := handler.NewMetricsHandler(svc)
	router := setupRouter(h)

	values := []string{"5", "3", "2"}
	for _, val := range values {
		req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/"+val, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/value/counter/PollCount", nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
	assert.Contains(t, wGet.Body.String(), "10")
}

func TestMetricsHandler_BatchUpdateAndGet(t *testing.T) {
	svc := newTestService(t)
	h := handler.NewMetricsHandler(svc)
	router := setupRouter(h)

	batch := []model.Metrics{
		{
			ID:    "Alloc",
			MType: model.Gauge,
			Value: ptrFloat64(555.55),
		},
		{
			ID:    "PollCount",
			MType: model.Counter,
			Delta: ptrInt64(99),
		},
	}

	body, err := json.Marshal(batch)
	require.NoError(t, err)

	reqUpdate := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	wUpdate := httptest.NewRecorder()
	router.ServeHTTP(wUpdate, reqUpdate)
	assert.Equal(t, http.StatusOK, wUpdate.Code)

	// Verify через JSON GET
	getMetric := model.Metrics{
		ID:    "Alloc",
		MType: model.Gauge,
	}
	getBody, err := json.Marshal(getMetric)
	require.NoError(t, err)

	reqGet := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(getBody))
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)

	var result model.Metrics
	err = json.NewDecoder(wGet.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, 555.55, *result.Value)
}

// === Вспомогательные функции ===

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
