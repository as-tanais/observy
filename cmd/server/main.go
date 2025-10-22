package main

import (
	"flag"
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

	serverAddr := flag.String("a", "localhost:8080", "Server address (host:port)")
	flag.Parse()

	storage := repository.NewMemStorage()
	service := service.NewMetricsService(storage)
	metricshandler := handler.NewMetricsHandler(service)

	router := chi.NewRouter()

	router.Post("/update/{type}/{name}/{value}", metricshandler.UpdateMetricHandler)
	router.Get("/value/{type}/{name}", metricshandler.GetMetricHandler)
	router.Get("/", metricshandler.ListMetricsHandler)

	fmt.Printf("Starting server on %s", *serverAddr)

	return http.ListenAndServe(*serverAddr, router)
}
