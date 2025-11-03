package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	fileStorage := repository.NewFileStorage(cfg.FileStoragePath)

	service := service.NewMetricsService(storage, fileStorage, cfg.StoreInterval)
	metricshandler := handler.NewMetricsHandler(service)

	if cfg.Restore {
		if err := service.LoadMetrics(); err != nil {
			logger.Warn("Failed to load metrics from file", zap.Error(err))
			logger.Info("Continuing without backup data")
		}
	}

	if cfg.StoreInterval > 0 {
		go startPeriodicSave(service, cfg.StoreInterval, logger)
	}

	router := chi.NewRouter()

	router.Use(middleware.WithLogging(logger))
	router.Use(middleware.GzipDecompressRequest())
	router.Use(middleware.GzipCompressResponse())

	router.Post("/update/", metricshandler.UpdateHandler)
	router.Post("/value/", metricshandler.GetMetric)

	router.Post("/update/{type}/{name}/{value}", metricshandler.UpdateMetricHandler)
	router.Get("/value/{type}/{name}", metricshandler.GetMetricHandler)

	router.Get("/", metricshandler.ListMetricsHandler)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("Server is ready", zap.String("listening_on", cfg.Address))
		serverErr <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server failed: %w", err)
		}
	case <-shutdown:
		logger.Info("Shutdown signal received, stopping server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Failed to gracefully shutdown server", zap.Error(err))
	} else {
		logger.Info("Server stopped")
	}

	if cfg.FileStoragePath != "" {
		if err := service.SaveToFile(); err != nil {
			logger.Error("Failed to save metrics on shutdown", zap.Error(err))
		} else {
			logger.Info("Metrics saved successfully on shutdown")
		}
	}

	return nil
}

func startPeriodicSave(service *service.MetricsService, interval time.Duration, logger *zap.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := service.SaveToFile(); err != nil {
			logger.Warn("Failed to save metrics", zap.Error(err))
		}
	}
}
