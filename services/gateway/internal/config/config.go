package config

import (
	"fmt"
	"os"
)

// Config holds everything the gateway needs to start.
type Config struct {
	Host string
	Port string

	// AuthServiceURL is where auth requests are forwarded.
	AuthServiceURL string

	// FarmServiceURL is where device and plant requests are forwarded.
	FarmServiceURL string
}

// Load reads config from the environment with defaults for local dev.
func Load() Config {
	return Config{
		Host:           getEnv("GATEWAY_HOST", "0.0.0.0"),
		Port:           getEnv("GATEWAY_PORT", "8000"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		FarmServiceURL: getEnv("FARM_SERVICE_URL", "http://localhost:8002"),
	}
}

// Addr returns the address the gateway listens on.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
