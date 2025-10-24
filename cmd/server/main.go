package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/as-tanais/observy/internal/config"
	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
	"github.com/as-tanais/observy/pkg/logger"
	"github.com/as-tanais/observy/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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

	logger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	defer logger.Sync()

	storage := repository.NewMemStorage()
	service := service.NewMetricsService(storage)
	metricshandler := handler.NewMetricsHandler(service)

	router := chi.NewRouter()

	router.Use(middleware.WithLogging(logger))

	router.Post("/update", metricshandler.UpdateHandler)
	router.Post("/value", metricshandler.GetMetric)

	router.Post("/update/{type}/{name}/{value}", metricshandler.UpdateMetricHandler)
	router.Get("/value/{type}/{name}", metricshandler.GetMetricHandler)

	router.Get("/", metricshandler.ListMetricsHandler)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	logger.Info("Server is ready",
		zap.String("listening_on", cfg.Address),
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
