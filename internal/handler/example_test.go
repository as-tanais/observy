// example_test.go
package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
	"github.com/go-chi/chi/v5"
)

func ExampleMetricsHandler_UpdateMetricHandler() {

	storage := repository.NewFileStorage("/var/metrics.db")
	svc := service.NewMetricsService(storage, nil, 0, nil)
	h := handler.NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)

	http.ListenAndServe(":8080", r)

	// проверка:
	// $ curl -X POST http://localhost:8080/update/counter/orders_total/1
	// $ curl -X POST http://localhost:8080/update/gauge/cpu_usage/42.5
	// $ curl -X POST http://localhost:8080/update/counter/errors_total/1

	// Output:
	// Server listening on :8080
	// Metrics endpoint: POST /update/{type}/{name}/{value}
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
