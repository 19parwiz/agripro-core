package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/19parwiz/agripro-core/services/auth/internal/config"
	"github.com/19parwiz/agripro-core/services/auth/internal/handler"
	"github.com/19parwiz/agripro-core/services/auth/internal/repository"
	"github.com/19parwiz/agripro-core/services/auth/internal/service"
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

	userRepo := repository.NewUserRepository(pool)
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTExpiry)
	authService := service.NewAuthService(userRepo, jwtManager)

	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(authService)

	router := newRouter(logger, healthHandler, authHandler)

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run the server in the background so we can listen for shutdown below.
	go func() {
		logger.Info("auth service listening", "addr", cfg.Addr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// Block until Ctrl+C or a container stop signal arrives.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down auth service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("auth service stopped")
}

func newRouter(logger *slog.Logger, health *handler.HealthHandler, auth *handler.AuthHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sharedmiddleware.RequestLogger(logger))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", health.Health)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
		r.Post("/verify-email", auth.VerifyEmail)
	})

	return r
}
