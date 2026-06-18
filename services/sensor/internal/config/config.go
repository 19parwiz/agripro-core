package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds everything the sensor service needs to start.
type Config struct {
	Host string
	Port string

	DatabaseURL string
	JWTSecret   string

	SnapshotURL       string
	IngestionEnabled  bool
	IngestDeviceID    string
	IngestionInterval time.Duration
}

// Load reads config from the environment with defaults for local dev.
func Load() Config {
	return Config{
		Host:        getEnv("SENSOR_HOST", "0.0.0.0"),
		Port:        getEnv("SENSOR_PORT", "8003"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),

		SnapshotURL:       getEnv("APP_SENSOR_SNAPSHOT_URL", "https://sensors.178-88-115-9.nip.io/data"),
		IngestionEnabled:  getEnvBool("APP_SENSOR_INGESTION_ENABLED", true),
		IngestDeviceID:    getEnv("APP_SENSOR_INGEST_DEVICE_ID", ""),
		IngestionInterval: getEnvDuration("APP_SENSOR_INGESTION_INTERVAL", 15*time.Minute),
	}
}

// Addr returns the address the sensor service listens on.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
