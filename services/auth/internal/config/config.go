package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds everything the auth service needs to start.
// Values come from environment variables so Railway/Docker can inject them.
type Config struct {
	Host string
	Port string

	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration
}

// Load reads config from the environment and applies sensible defaults for local dev.
func Load() Config {
	return Config{
		Host:        getEnv("AUTH_HOST", "0.0.0.0"),
		Port:        getEnv("AUTH_PORT", "8001"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:   time.Duration(getEnvInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
	}
}

// Addr returns the address the HTTP server should listen on.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
