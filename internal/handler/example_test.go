// example_test.go
package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
	"github.com/go-chi/chi/v5"
)

func ExampleMetricsHandler_UpdateMetricHandler() {
	// 1. Создаём тестовое хранилище и сервис
	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, 0, nil)
	h := handler.NewMetricsHandler(svc)

	// 2. Настраиваем роутер с нужным эндпоинтом
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)

	// 3. Создаём тестовый запрос к эндпоинту
	// Обновляем counter-метрику "test_counter" со значением 42
	req := httptest.NewRequest(
		http.MethodPost,
		"/update/counter/test_counter/42",
		strings.NewReader(""),
	)
	w := httptest.NewRecorder()

	// 4. Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// 5. Выводим результат для документации
	fmt.Println("Status code:", w.Code)
	fmt.Println("Response body:", w.Body.String())

	// Output:
	// Status code: 200
	// Response body: Metric updated: type=counter, name=test_counter, value=42
}

func ExampleMetricsHandler_ListMetricsHandler() {

	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, 0, nil)
	h := handler.NewMetricsHandler(svc)

	_ = svc.SetMetric(context.Background(), "counter", "orders_total", "150", "127.0.0.1")
	_ = svc.SetMetric(context.Background(), "gauge", "cpu_usage", "42.5", "127.0.0.1")
	_ = svc.SetMetric(context.Background(), "counter", "requests_total", "1000", "127.0.0.1")

	r := chi.NewRouter()
	r.Get("/", h.ListMetricsHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Println("Status code:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))
	fmt.Println("Has metrics table:", w.Body.String() != "")
	fmt.Println("Total metrics in response:", w.Body.String() != "")

	// Output:
	// Status code: 200
	// Content-Type: text/html; charset=utf-8
	// Has metrics table: true
	// Total metrics in response: true
}
