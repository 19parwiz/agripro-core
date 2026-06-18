package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/19parwiz/agripro-core/services/sensor/internal/model"
)

type snapshotReading struct {
	SensorType string
	Value      float64
	Unit       string
}

type sensorField struct {
	sensorType string
	unit       string
	keys       []string
}

var snapshotFields = []sensorField{
	{sensorType: model.SensorTemperature, unit: "°C", keys: []string{"temperature", "air_temperature", "temp"}},
	{sensorType: model.SensorHumidity, unit: "%", keys: []string{"humidity", "air_humidity"}},
	{sensorType: model.SensorPH, unit: "pH", keys: []string{"ph", "pH"}},
	{sensorType: model.SensorEC, unit: "mS/cm", keys: []string{"ec", "EC"}},
	{sensorType: model.SensorCO2, unit: "ppm", keys: []string{"co2", "CO2"}},
	{sensorType: model.SensorTDS, unit: "ppm", keys: []string{"tds", "TDS"}},
}

func parseSnapshot(body []byte) ([]snapshotReading, time.Time, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, time.Time{}, fmt.Errorf("decode snapshot json: %w", err)
	}

	recordedAt := parseSnapshotTimestamp(raw)

	var readings []snapshotReading
	for _, field := range snapshotFields {
		for _, key := range field.keys {
			msg, ok := raw[key]
			if !ok {
				continue
			}

			value, unit, found := parseSensorValue(msg, field.unit)
			if !found {
				continue
			}

			readings = append(readings, snapshotReading{
				SensorType: field.sensorType,
				Value:      value,
				Unit:       unit,
			})
			break
		}
	}

	if len(readings) == 0 {
		return nil, recordedAt, fmt.Errorf("snapshot has no known sensor fields")
	}

	return readings, recordedAt, nil
}

func parseSnapshotTimestamp(raw map[string]json.RawMessage) time.Time {
	now := time.Now().UTC()

	for _, key := range []string{"timestamp", "recorded_at", "time"} {
		msg, ok := raw[key]
		if !ok {
			continue
		}

		var value string
		if err := json.Unmarshal(msg, &value); err == nil {
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				return parsed.UTC()
			}
		}

		var unixSeconds float64
		if err := json.Unmarshal(msg, &unixSeconds); err == nil {
			return time.Unix(int64(unixSeconds), 0).UTC()
		}
	}

	return now
}

func parseSensorValue(msg json.RawMessage, defaultUnit string) (float64, string, bool) {
	var nested struct {
		Value json.Number `json:"value"`
		Unit  string      `json:"unit"`
	}
	if err := json.Unmarshal(msg, &nested); err == nil && nested.Value != "" {
		value, err := nested.Value.Float64()
		if err != nil {
			return 0, "", false
		}

		unit := strings.TrimSpace(nested.Unit)
		if unit == "" {
			unit = defaultUnit
		}

		return value, unit, true
	}

	var number float64
	if err := json.Unmarshal(msg, &number); err == nil {
		return number, defaultUnit, true
	}

	var text string
	if err := json.Unmarshal(msg, &text); err == nil {
		value, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil {
			return 0, "", false
		}
		return value, defaultUnit, true
	}

	return 0, "", false
}
