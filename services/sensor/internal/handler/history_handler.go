package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/19parwiz/agripro-core/services/sensor/internal/service"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	sharedmiddleware "github.com/19parwiz/agripro-core/shared/middleware"
	"github.com/19parwiz/agripro-core/shared/response"
)

type HistoryHandler struct {
	sensors *service.SensorService
}

func NewHistoryHandler(sensors *service.SensorService) *HistoryHandler {
	return &HistoryHandler{sensors: sensors}
}

type historyResponse struct {
	DeviceID   string                 `json:"device_id"`
	SensorType string                 `json:"sensor_type"`
	From       time.Time              `json:"from"`
	To         time.Time              `json:"to"`
	Points     []service.HistoryPoint `json:"points"`
}

// Hourly returns hourly chart buckets for one device and sensor type.
func (h *HistoryHandler) Hourly(w http.ResponseWriter, r *http.Request) {
	h.handleHistory(w, r, 24*time.Hour, h.sensors.HourlyHistory)
}

// Daily returns daily chart buckets for one device and sensor type.
func (h *HistoryHandler) Daily(w http.ResponseWriter, r *http.Request) {
	h.handleHistory(w, r, 30*24*time.Hour, h.sensors.DailyHistory)
}

func (h *HistoryHandler) handleHistory(
	w http.ResponseWriter,
	r *http.Request,
	defaultRange time.Duration,
	load func(context.Context, string, string, string, time.Time, time.Time) ([]service.HistoryPoint, error),
) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	deviceID, err := requiredQuery(r, "device_id")
	if err != nil {
		writeError(w, err)
		return
	}

	sensorType, err := requiredQuery(r, "sensor_type")
	if err != nil {
		writeError(w, err)
		return
	}

	from, to, err := parseHistoryRange(r, defaultRange)
	if err != nil {
		writeError(w, err)
		return
	}

	points, err := load(r.Context(), claims.UserID, deviceID, sensorType, from, to)
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "sensor history loaded", historyResponse{
		DeviceID:   deviceID,
		SensorType: sensorType,
		From:       from,
		To:         to,
		Points:     points,
	})
}
