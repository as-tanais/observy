package agent

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"sync/atomic"
// 	"testing"

// 	models "github.com/as-tanais/observy/internal/model"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestCollect(t *testing.T) {

// 	metrics := Collect()

// 	require.NotEmpty(t, metrics, "Метрики не должны быть пустыми")

// 	expectedCount := 29 //Все метрики из задания(библиотека runtime) + 2 доп
// 	assert.Equal(t, expectedCount, len(metrics), "Должно быть собрано 29 метрик")

// 	for _, metric := range metrics {
// 		assert.NotEmpty(t, metric.ID, "ID метрики не должен быть пустым")
// 		assert.NotEmpty(t, metric.MType, "Тип метрики не должен быть пустым")
// 	}
// }

// func TestCollect_PollCountIncrement(t *testing.T) {
// 	atomic.StoreInt64(&pollCount, 0)

// 	for i := 1; i <= 3; i++ {
// 		metrics := Collect()

// 		var found bool
// 		var pollCountValue int64

// 		for _, metric := range metrics {
// 			if metric.ID == "PollCount" {
// 				found = true
// 				require.NotNil(t, metric.Delta, "PollCount Delta не должна быть nil")
// 				pollCountValue = *metric.Delta
// 				break
// 			}
// 		}

// 		assert.True(t, found, "Метрика PollCount должна присутствовать")

// 		assert.Equal(t, int64(i), pollCountValue, "PollCount должен быть равен %d после %d вызова", i, i)
// 	}
// }

// func TestSend_URLFormat(t *testing.T) {
// 	var capturedPath string

// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		capturedPath = r.URL.Path
// 		w.WriteHeader(http.StatusOK)
// 	}))
// 	defer server.Close()

// 	value := 123.45
// 	testMetrics := []models.Metrics{
// 		{ID: "TestGauge", MType: models.Gauge, Value: &value},
// 	}

// 	Send(testMetrics, server.URL)

// 	expectedPath := "/update/gauge/TestGauge/123.450000"
// 	assert.Equal(t, expectedPath, capturedPath, "URL должен соответствовать формату /update/{type}/{name}/{value}")
// }

// func TestSend_CounterMetric(t *testing.T) {
// 	var capturedPath string
// 	var capturedMethod string
// 	var capturedContentType string

// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		capturedPath = r.URL.Path
// 		capturedMethod = r.Method
// 		capturedContentType = r.Header.Get("Content-Type")
// 		w.WriteHeader(http.StatusOK)
// 	}))
// 	defer server.Close()

// 	delta := int64(42)
// 	testMetrics := []models.Metrics{
// 		{ID: "PollCount", MType: models.Counter, Delta: &delta},
// 	}

// 	Send(testMetrics, server.URL)

// 	expectedPath := "/update/counter/PollCount/42"
// 	assert.Equal(t, expectedPath, capturedPath, "Путь должен содержать counter тип и целое значение")

// 	assert.Equal(t, http.MethodPost, capturedMethod, "Метод должен быть POST")

// 	assert.Equal(t, "text/plain", capturedContentType, "Content-Type должен быть text/plain")
// }

// func TestSend_MultipleMetrics(t *testing.T) {
// 	requestCount := 0
// 	receivedPaths := make([]string, 0)

// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		requestCount++
// 		receivedPaths = append(receivedPaths, r.URL.Path)
// 		w.WriteHeader(http.StatusOK)
// 	}))
// 	defer server.Close()

// 	gaugeValue := 100.5
// 	counterDelta := int64(5)

// 	testMetrics := []models.Metrics{
// 		{ID: "Alloc", MType: models.Gauge, Value: &gaugeValue},
// 		{ID: "PollCount", MType: models.Counter, Delta: &counterDelta},
// 		{ID: "Sys", MType: models.Gauge, Value: &gaugeValue},
// 	}

// 	Send(testMetrics, server.URL)

// 	assert.Equal(t, 3, requestCount, "Должно быть отправлено 3 запроса")

// 	expectedPaths := []string{
// 		"/update/gauge/Alloc/100.500000",
// 		"/update/counter/PollCount/5",
// 		"/update/gauge/Sys/100.500000",
// 	}

// 	for _, expectedPath := range expectedPaths {
// 		assert.Contains(t, receivedPaths, expectedPath, "Должен быть отправлен запрос на путь %s", expectedPath)
// 	}
// }
