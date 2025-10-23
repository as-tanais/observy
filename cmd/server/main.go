package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/as-tanais/observy/internal/config"
	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	storage := repository.NewMemStorage()
	service := service.NewMetricsService(storage)
	metricshandler := handler.NewMetricsHandler(service)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", metricshandler.UpdateMetricHandler)
	router.Get("/value/{type}/{name}", metricshandler.GetMetricHandler)
	router.Get("/", metricshandler.ListMetricsHandler)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	fmt.Printf("Starting server on %s\n", cfg.Address)
	log.Printf("Listening on: %s", cfg.Address)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
