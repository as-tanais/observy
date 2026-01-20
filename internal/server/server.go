package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/as-tanais/observy/internal/audit"
	"github.com/as-tanais/observy/internal/config"
	"github.com/as-tanais/observy/internal/handler"
	"github.com/as-tanais/observy/internal/repository"
	"github.com/as-tanais/observy/internal/service"
	"github.com/as-tanais/observy/pkg/helpers/pg"
	"github.com/as-tanais/observy/pkg/logger"
	"github.com/as-tanais/observy/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func Run() error {
	cfg, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Sync()

	ctx := context.Background()

	var pool *pgxpool.Pool
	var storage repository.Storage
	var fileStorage *repository.FileStorage

	var usedStorage string

	if cfg.DSN != "" {
		if err := dbMigrate(cfg.DSN); err != nil {
			log.Fatal("Migration failed", zap.Error(err))
		}

		dbConfig, err := pg.NewPoolConfig(cfg.DSN)
		if err != nil {
			log.Fatal("Parse DB config failed", zap.Error(err))
		}

		pool, err = pg.NewConnection(ctx, dbConfig)
		if err != nil {
			log.Fatal("DB connection failed", zap.Error(err))
		}

		storage = repository.NewPGStorage(pool)
		usedStorage = "Postgres DB"
	} else if cfg.FileStoragePath != "" {
		storage = repository.NewFileStorage(cfg.FileStoragePath)
		usedStorage = "File"
	} else {
		storage = repository.NewMemStorage()
		usedStorage = "Mem storage"
	}

	if cfg.FileStoragePath != "" {
		fileStorage = repository.NewFileStorage(cfg.FileStoragePath)
	}

	var observer *audit.Observer
	if cfg.AuditFile != "" || cfg.AuditURL != "" {
		var subs []audit.Subscriber

		if cfg.AuditFile != "" {
			fileSub, err := audit.NewFileSub(cfg.AuditFile)
			if err != nil {
				log.Warn("Не удалось создать файловый адитор", zap.Error(err))
			} else {
				subs = append(subs, fileSub)
			}
		}

		if cfg.AuditURL != "" {
			httpSub := audit.NewHTTPSub(cfg.AuditURL)
			subs = append(subs, httpSub)
		}

		if len(subs) > 0 {
			observer = audit.NewObserver(subs...)
		}
	}

	service := service.NewMetricsService(storage, fileStorage, cfg.StoreInterval, observer)

	if cfg.StoreInterval > 0 && fileStorage != nil {
		go startPeriodicSave(service, cfg.StoreInterval, log)
	}

	if cfg.Restore && fileStorage != nil {
		if err := service.LoadMetrics(ctx); err != nil {
			log.Warn("Failed to load metrics from file", zap.Error(err))
		}
	}

	router := chi.NewRouter()
	router.Use(middleware.WithLogging(log))
	router.Use(middleware.GzipDecompressRequest())
	if cfg.Key != "" {
		router.Use(middleware.SignatureMiddleware(cfg.Key))
	}
	router.Use(middleware.GzipCompressResponse())

	h := handler.NewMetricsHandler(service)

	router.Post("/update/", h.UpdateHandler)
	router.Post("/value/", h.GetMetric)
	router.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	router.Get("/value/{type}/{name}", h.GetMetricHandler)
	router.Get("/", h.ListMetricsHandler)
	router.Post("/updates/", h.UpdateMetricsHandler)

	if pool != nil {
		router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			if err := pool.Ping(r.Context()); err != nil {
				log.Warn("Database ping failed", zap.Error(err))
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
	}

	server := &http.Server{Addr: cfg.Address, Handler: router}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("Server is ready", zap.String("listening on", cfg.Address))
		log.Info("Server use", zap.String("storage", usedStorage))
		serverErr <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	var errServer error
	select {
	case errServer = <-serverErr:
		if errServer != nil && errServer != http.ErrServerClosed {
			return fmt.Errorf("server failed: %w", errServer)
		}
	case <-shutdown:
		log.Info("Shutdown signal received, stopping server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Failed to gracefully shutdown server", zap.Error(err))
	} else {
		log.Info("Server stopped")
	}

	if fileStorage != nil {
		if err := service.SaveToFile(ctx); err != nil {
			log.Error("Failed to save metrics on shutdown", zap.Error(err))
		} else {
			log.Info("Metrics saved successfully on shutdown")
		}
	}

	return nil
}

func startPeriodicSave(svc *service.MetricsService, interval time.Duration, logger *zap.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx := context.Background()

	for range ticker.C {
		if err := svc.SaveToFile(ctx); err != nil {
			logger.Warn("Failed to save metrics", zap.Error(err))
		}
	}
}
