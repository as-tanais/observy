package main

import (
	"net/http"

	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	storage := repository.NewMemStorage()
	service := service.NewMetricsService(storage)
	metricshandler := handler.NewMetricsHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/update/", metricshandler.UpdateMetricHandler)
	mux.HandleFunc("/value/", metricshandler.GetMetricHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return http.ListenAndServe(`:8080`, mux)
}
