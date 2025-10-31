package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/as-tanais/observy/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollect(t *testing.T) {
	metrics := Collect()

	require.NotEmpty(t, metrics, "Метрики не должны быть пустыми")

	expectedCount := 29
	assert.Equal(t, expectedCount, len(metrics), "Должно быть собрано 29 метрик")

	for _, metric := range metrics {
		assert.NotEmpty(t, metric.ID, "ID метрики не должен быть пустым")
		assert.NotEmpty(t, metric.MType, "Тип метрики не должен быть пустым")
	}
}

func TestCollect_PollCountIncrement(t *testing.T) {
	for i := 1; i <= 3; i++ {
		metrics := Collect()

		var found bool
		var pollCountDelta int64

		for _, metric := range metrics {
			if metric.ID == "PollCount" {
				found = true
				require.NotNil(t, metric.Delta, "PollCount Delta не должна быть nil")
				pollCountDelta = *metric.Delta
				break
			}
		}

		assert.True(t, found, "Метрика PollCount должна присутствовать")

		assert.Equal(t, int64(1), pollCountDelta, "PollCount delta должна быть 1 при каждом вызове")
	}
}

func TestSend_JSONFormat(t *testing.T) {
	var capturedPath string
	var capturedMethod string
	var capturedContentType string
	var capturedContentEncoding string
	var capturedBody models.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		capturedContentType = r.Header.Get("Content-Type")
		capturedContentEncoding = r.Header.Get("Content-Encoding")

		// Читаем тело запроса
		body, _ := io.ReadAll(r.Body)

		// Если тело сжато - распаковываем
		if capturedContentEncoding == "gzip" {
			gr, err := gzip.NewReader(bytes.NewReader(body))
			if err == nil {
				defer gr.Close()
				body, _ = io.ReadAll(gr)
			}
		}

		json.Unmarshal(body, &capturedBody)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	value := 123.45
	testMetrics := []models.Metrics{
		{ID: "TestGauge", MType: models.Gauge, Value: &value},
	}

	Send(testMetrics, server.URL)

	assert.Equal(t, "/update/", capturedPath, "URL должен быть /update/ (JSON API)")
	assert.Equal(t, http.MethodPost, capturedMethod, "Метод должен быть POST")
	assert.Equal(t, "application/json", capturedContentType, "Content-Type должен быть application/json")
	assert.Equal(t, "gzip", capturedContentEncoding, "Content-Encoding должен быть gzip")

	assert.Equal(t, "TestGauge", capturedBody.ID, "ID в JSON должен быть TestGauge")
	assert.Equal(t, models.Gauge, capturedBody.MType, "Type в JSON должен быть gauge")
	assert.NotNil(t, capturedBody.Value, "Value не должен быть nil")
	assert.Equal(t, value, *capturedBody.Value, "Value должен соответствовать отправленному")
}

func TestSend_CounterMetric(t *testing.T) {
	var capturedPath string
	var capturedMethod string
	var capturedContentType string
	var capturedContentEncoding string
	var capturedBody models.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedMethod = r.Method
		capturedContentType = r.Header.Get("Content-Type")
		capturedContentEncoding = r.Header.Get("Content-Encoding")

		body, _ := io.ReadAll(r.Body)

		// Если тело сжато - распаковываем
		if capturedContentEncoding == "gzip" {
			gr, err := gzip.NewReader(bytes.NewReader(body))
			if err == nil {
				defer gr.Close()
				body, _ = io.ReadAll(gr)
			}
		}

		json.Unmarshal(body, &capturedBody)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	delta := int64(1)
	testMetrics := []models.Metrics{
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
	}

	Send(testMetrics, server.URL)

	assert.Equal(t, "/update/", capturedPath, "Путь должен быть /update/ (JSON API)")
	assert.Equal(t, http.MethodPost, capturedMethod, "Метод должен быть POST")
	assert.Equal(t, "application/json", capturedContentType, "Content-Type должен быть application/json")
	assert.Equal(t, "gzip", capturedContentEncoding, "Content-Encoding должен быть gzip")

	// Проверяем JSON содержимое
	assert.Equal(t, "PollCount", capturedBody.ID, "ID в JSON должен быть PollCount")
	assert.Equal(t, models.Counter, capturedBody.MType, "Type в JSON должен быть counter")
	assert.NotNil(t, capturedBody.Delta, "Delta не должен быть nil")
	assert.Equal(t, int64(1), *capturedBody.Delta, "Delta должен быть 1")
}

func TestSend_MultipleMetrics(t *testing.T) {
	requestCount := 0
	receivedMetrics := make([]models.Metrics, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		body, _ := io.ReadAll(r.Body)

		// Если тело сжато - распаковываем
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(bytes.NewReader(body))
			if err == nil {
				defer gr.Close()
				body, _ = io.ReadAll(gr)
			}
		}

		var metric models.Metrics
		json.Unmarshal(body, &metric)
		receivedMetrics = append(receivedMetrics, metric)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	gaugeValue := 100.5
	counterDelta := int64(1)

	testMetrics := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &gaugeValue},
		{ID: "PollCount", MType: models.Counter, Delta: &counterDelta},
		{ID: "Sys", MType: models.Gauge, Value: &gaugeValue},
	}

	Send(testMetrics, server.URL)

	assert.Equal(t, 3, requestCount, "Должно быть отправлено 3 запроса")

	metricNames := make([]string, 0)
	for _, m := range receivedMetrics {
		metricNames = append(metricNames, m.ID)
	}

	assert.Contains(t, metricNames, "Alloc", "Должна быть отправлена метрика Alloc")
	assert.Contains(t, metricNames, "PollCount", "Должна быть отправлена метрика PollCount")
	assert.Contains(t, metricNames, "Sys", "Должна быть отправлена метрика Sys")
}

func TestSend_ErrorHandling(t *testing.T) {
	testMetrics := []models.Metrics{
		{ID: "TestMetric", MType: models.Gauge, Value: ptrFloat64(1.0)},
	}

	assert.NotPanics(t, func() {
		Send(testMetrics, "http://localhost:9999")
	})
}

func TestValidateMetric(t *testing.T) {
	tests := []struct {
		name      string
		metric    models.Metrics
		wantError bool
	}{
		{
			name:      "valid gauge",
			metric:    models.Metrics{ID: "Test", MType: models.Gauge, Value: ptrFloat64(1.0)},
			wantError: false,
		},
		{
			name:      "valid counter",
			metric:    models.Metrics{ID: "Test", MType: models.Counter, Delta: ptrInt64(1)},
			wantError: false,
		},
		{
			name:      "empty ID",
			metric:    models.Metrics{ID: "", MType: models.Gauge, Value: ptrFloat64(1.0)},
			wantError: true,
		},
		{
			name:      "gauge with nil value",
			metric:    models.Metrics{ID: "Test", MType: models.Gauge, Value: nil},
			wantError: true,
		},
		{
			name:      "counter with nil delta",
			metric:    models.Metrics{ID: "Test", MType: models.Counter, Delta: nil},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetric(tt.metric)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
