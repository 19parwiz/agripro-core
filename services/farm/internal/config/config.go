package config

import (
	"fmt"
	"os"
)

// Config holds everything the farm service needs to start.
type Config struct {
	Host string
	Port string

	DatabaseURL string
	JWTSecret   string
}

// Load reads config from the environment with defaults for local dev.
func Load() Config {
	return Config{
		Host:        getEnv("FARM_HOST", "0.0.0.0"),
		Port:        getEnv("FARM_PORT", "8002"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
	}
}

// Addr returns the address the farm service listens on.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
