package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/19parwiz/agripro-core/services/sensor/internal/config"
	"github.com/19parwiz/agripro-core/services/sensor/internal/handler"
	"github.com/19parwiz/agripro-core/services/sensor/internal/repository"
	"github.com/19parwiz/agripro-core/services/sensor/internal/service"
	"github.com/19parwiz/agripro-core/shared/jwt"
	sharedmiddleware "github.com/19parwiz/agripro-core/shared/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	appCtx, stopSchedulers := context.WithCancel(context.Background())
	defer stopSchedulers()

	pool, err := repository.NewPool(appCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	deviceRepo := repository.NewDeviceRepository(pool)
	readingRepo := repository.NewReadingRepository(pool)
	archiveRepo := repository.NewArchiveRepository(pool)

	sensorService := service.NewSensorService(cfg, deviceRepo, readingRepo, archiveRepo)
	sensorService.StartSchedulers(appCtx)

	jwtManager := jwt.NewManager(cfg.JWTSecret, 24*time.Hour)

	healthHandler := handler.NewHealthHandler()
	historyHandler := handler.NewHistoryHandler(sensorService)

	router := newRouter(logger, jwtManager, healthHandler, historyHandler)

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("sensor service listening", "addr", cfg.Addr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down sensor service")
	stopSchedulers()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("sensor service stopped")
}

func newRouter(
	logger *slog.Logger,
	jwtManager *jwt.Manager,
	health *handler.HealthHandler,
	history *handler.HistoryHandler,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sharedmiddleware.RequestLogger(logger))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", health.Health)

	r.Route("/api", func(r chi.Router) {
		r.Use(sharedmiddleware.Authenticate(jwtManager))

		r.Route("/sensors", func(r chi.Router) {
			r.Route("/history", func(r chi.Router) {
				r.Get("/hourly", history.Hourly)
				r.Get("/daily", history.Daily)
			})
		})
	})

	return r
}
