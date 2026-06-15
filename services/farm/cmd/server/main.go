package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/19parwiz/agripro-core/services/farm/internal/config"
	"github.com/19parwiz/agripro-core/services/farm/internal/handler"
	"github.com/19parwiz/agripro-core/services/farm/internal/repository"
	"github.com/19parwiz/agripro-core/services/farm/internal/service"
	sharedmiddleware "github.com/19parwiz/agripro-core/shared/middleware"
	"github.com/19parwiz/agripro-core/shared/jwt"
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

	ctx := context.Background()

	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	deviceRepo := repository.NewDeviceRepository(pool)
	plantRepo := repository.NewPlantRepository(pool)

	deviceService := service.NewDeviceService(deviceRepo)
	plantService := service.NewPlantService(plantRepo)

	jwtManager := jwt.NewManager(cfg.JWTSecret, 24*time.Hour)

	healthHandler := handler.NewHealthHandler()
	deviceHandler := handler.NewDeviceHandler(deviceService)
	plantHandler := handler.NewPlantHandler(plantService)

	router := newRouter(logger, jwtManager, healthHandler, deviceHandler, plantHandler)

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("farm service listening", "addr", cfg.Addr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down farm service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("farm service stopped")
}

func newRouter(
	logger *slog.Logger,
	jwtManager *jwt.Manager,
	health *handler.HealthHandler,
	devices *handler.DeviceHandler,
	plants *handler.PlantHandler,
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

		r.Route("/devices", func(r chi.Router) {
			r.Post("/", devices.Create)
			r.Get("/", devices.List)
			r.Get("/{id}", devices.Get)
			r.Put("/{id}", devices.Update)
			r.Delete("/{id}", devices.Delete)
		})

		r.Route("/plants", func(r chi.Router) {
			r.Post("/", plants.Create)
			r.Get("/", plants.List)
			r.Get("/{id}", plants.Get)
			r.Put("/{id}", plants.Update)
			r.Delete("/{id}", plants.Delete)
		})
	})

	return r
}
