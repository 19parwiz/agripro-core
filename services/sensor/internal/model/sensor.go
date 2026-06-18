package model

import "time"

// Common sensor types from the lab Pi snapshot.
const (
	SensorTemperature = "temperature"
	SensorHumidity    = "humidity"
	SensorPH          = "ph"
	SensorEC          = "ec"
	SensorCO2         = "co2"
	SensorTDS         = "tds"
)

// Reading is one raw sensor value stored after ingestion.
type Reading struct {
	ID         string
	DeviceID   string
	SensorType string
	Value      float64
	Unit       string
	RecordedAt time.Time
	CreatedAt  time.Time
}

// HourlyArchive holds aggregated stats for one device, type, and hour.
type HourlyArchive struct {
	ID           string
	DeviceID     string
	SensorType   string
	BucketStart  time.Time
	AvgValue     float64
	MinValue     float64
	MaxValue     float64
	ReadingCount int
	CreatedAt    time.Time
}

// DailyArchive holds aggregated stats for one device, type, and day.
type DailyArchive struct {
	ID           string
	DeviceID     string
	SensorType   string
	BucketDate   time.Time
	AvgValue     float64
	MinValue     float64
	MaxValue     float64
	ReadingCount int
	CreatedAt    time.Time
}
