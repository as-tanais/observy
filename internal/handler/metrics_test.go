package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/as-tanais/observy/internal/handler"
	model "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
)

func setupRouter(h *handler.MetricsHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Get("/value/{type}/{name}", h.GetMetricHandler)
	return r
}

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
			wantBody:       "",
		},
		{
			name:           "неправильный формат пути (не хватает значения)",
			method:         http.MethodPost,
			path:           "/update/gauge/Alloc",
			wantStatusCode: http.StatusNotFound,
			wantBody:       "",
		},
		{
			name:           "неправильный тип метрики",
			method:         http.MethodPost,
			path:           "/update/unknown/TestMetric/123",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "unknown metric type",
		},
		{
			name:           "невалидное значение для gauge",
			method:         http.MethodPost,
			path:           "/update/gauge/Alloc/notanumber",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid gauge value",
		},
		{
			name:           "невалидное значение для counter",
			method:         http.MethodPost,
			path:           "/update/counter/PollCount/notanumber",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "invalid counter value",
		},
		{
			name:           "пустое имя метрики",
			method:         http.MethodPost,
			path:           "/update/gauge//123",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage := repository.NewMemStorage()
			svc := service.NewMetricsService(storage)
			h := handler.NewMetricsHandler(svc)

			router := setupRouter(h)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code, "Неправильный статус код")

			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestMetricsHandler_GetMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMetrics   map[string]string
		method         string
		path           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "успешное получение gauge метрики",
			setupMetrics: map[string]string{
				"Alloc": "123.45",
			},
			method:         http.MethodGet,
			path:           "/value/gauge/Alloc",
			wantStatusCode: http.StatusOK,
			wantBody:       "123.45",
		},
		{
			name: "успешное получение counter метрики",
			setupMetrics: map[string]string{
				"PollCount": "42",
			},
			method:         http.MethodGet,
			path:           "/value/counter/PollCount",
			wantStatusCode: http.StatusOK,
			wantBody:       "42",
		},
		{
			name:           "метрика не найдена",
			setupMetrics:   map[string]string{},
			method:         http.MethodGet,
			path:           "/value/gauge/NonExistent",
			wantStatusCode: http.StatusNotFound,
			wantBody:       "Metric not found",
		},
		{
			name: "несоответствие типа метрики",
			setupMetrics: map[string]string{
				"PollCount": "10",
			},
			method:         http.MethodGet,
			path:           "/value/gauge/PollCount",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "Metric type mismatch",
		},
		{
			name:           "неправильный HTTP метод (POST вместо GET)",
			setupMetrics:   map[string]string{},
			method:         http.MethodPost,
			path:           "/value/gauge/Alloc",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "",
		},
		{
			name:           "неправильный формат пути",
			setupMetrics:   map[string]string{},
			method:         http.MethodGet,
			path:           "/value/gauge",
			wantStatusCode: http.StatusNotFound,
			wantBody:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage := repository.NewMemStorage()
			svc := service.NewMetricsService(storage)
			h := handler.NewMetricsHandler(svc)

			for name, value := range tt.setupMetrics {
				var metricType string
				if name == "PollCount" {
					metricType = model.Counter
				} else {
					metricType = model.Gauge
				}
				err := svc.SetMetric(metricType, name, value)
				require.NoError(t, err, "Ошибка при предзагрузке метрики")
			}

			router := setupRouter(h)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code, "Неправильный статус код")

			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestMetricsHandler_UpdateAndGet(t *testing.T) {
	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage)
	h := handler.NewMetricsHandler(svc)
	router := setupRouter(h)

	reqUpdate := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/999.99", nil)
	wUpdate := httptest.NewRecorder()
	router.ServeHTTP(wUpdate, reqUpdate)
	assert.Equal(t, http.StatusOK, wUpdate.Code)

	reqGet := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
	assert.Equal(t, "999.99", wGet.Body.String())
}

func TestMetricsHandler_CounterAccumulation(t *testing.T) {
	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage)
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
	assert.Equal(t, "10", wGet.Body.String())
}
