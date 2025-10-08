package main

import (
	"fmt"
	"net/http"

	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
	"github.com/go-chi/chi/v5"
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

	router := chi.NewRouter()

	router.Post("/update/{type}/{name}/{value}", metricshandler.UpdateMetricHandler)
	router.Get("/value/{type}/{name}", metricshandler.GetMetricHandler)

	fmt.Println("Starting server on :8080")

	return http.ListenAndServe(`:8080`, router)
}
