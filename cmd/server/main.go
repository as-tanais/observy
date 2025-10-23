package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

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
	serverAddr := flag.String("a", "localhost:8080", "Server address (host:port)")

	flag.Parse()

	addr := os.Getenv("ADDRESS")
	if addr == "" {
		addr = *serverAddr
	}

	storage := repository.NewMemStorage()
	service := service.NewMetricsService(storage)
	metricshandler := handler.NewMetricsHandler(service)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", metricshandler.UpdateMetricHandler)
	router.Get("/value/{type}/{name}", metricshandler.GetMetricHandler)
	router.Get("/", metricshandler.ListMetricsHandler)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
