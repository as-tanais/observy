// example_test.go
package handler_test

import (
	"fmt"

	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
	"github.com/go-chi/chi/v5"
)

func ExampleMetricsHandler_UpdateMetricHandler() {
	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, 0, nil)
	h := handler.NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)

	// Запускаем сервер :
	//   http.ListenAndServe(":8080", r)
	//
	// Примеры запросов:
	//   $ curl -X POST http://localhost:8080/update/counter/orders_total/1
	//   $ curl -X POST http://localhost:8080/update/gauge/cpu_usage/42.5

	fmt.Println("Handler registered at /update/{type}/{name}/{value}")
	// Output: Handler registered at /update/{type}/{name}/{value}
}

func ExampleMetricsHandler_ListMetricsHandler() {
	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, 0, nil)
	h := handler.NewMetricsHandler(svc)

	// Регистрируем хендлер для просмотра всех метрик
	r := chi.NewRouter()
	r.Get("/metrics", h.ListMetricsHandler)

	// Запускаем сервер (в реальном приложении):
	//   http.ListenAndServe(":8080", r)
	//
	// Теперь можно получить все метрики:
	//   $ curl http://localhost:8080/metrics
	//   <!DOCTYPE html>
	//   <html>
	//   <body>
	//     <h1>Metrics</h1>
	//     <table>
	//       <tr><th>Type</th><th>Name</th><th>Value</th></tr>
	//       <tr><td>counter</td><td>orders_total</td><td>150</td></tr>
	//       <tr><td>gauge</td><td>cpu_usage</td><td>42.5</td></tr>
	//     </table>
	//   </body>
	//   </html>

	fmt.Println("Handler registered at /metrics")
	// Output: Handler registered at /metrics
}

func ExampleMetricsHandler_GetMetric() {
	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, 0, nil)
	h := handler.NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Get("/value/", h.GetMetric)

	// Запускаем сервер (в реальном приложении):
	//   http.ListenAndServe(":8080", r)
	//
	// Теперь можно получить все метрики:
	//   $ curl http://localhost:8080/value/
	//   <!DOCTYPE html>
	//   <html>
	//   <body>
	//     <h1>Metrics</h1>
	//     <table>
	//       <tr><th>Type</th><th>Name</th><th>Value</th></tr>
	//       <tr><td>counter</td><td>orders_total</td><td>150</td></tr>
	//       <tr><td>gauge</td><td>cpu_usage</td><td>42.5</td></tr>
	//     </table>
	//   </body>
	//   </html>

	fmt.Println("Handler registered at /value/")
	// Output: Handler registered at /value/
}
