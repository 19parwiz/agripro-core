package handler

import (
	"net/http"

	"github.com/19parwiz/agripro-core/services/farm/internal/service"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	sharedmiddleware "github.com/19parwiz/agripro-core/shared/middleware"
	"github.com/19parwiz/agripro-core/shared/response"
	"github.com/go-chi/chi/v5"
)

type DeviceHandler struct {
	devices *service.DeviceService
}

func NewDeviceHandler(devices *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{devices: devices}
}

type deviceRequest struct {
	Name       string `json:"name"`
	DeviceID   string `json:"device_id"`
	Type       string `json:"type"`
	Location   string `json:"location"`
	StreamPath string `json:"stream_path"`
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	var req deviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.devices.Create(r.Context(), claims.UserID, service.DeviceInput{
		Name:       req.Name,
		DeviceID:   req.DeviceID,
		Type:       req.Type,
		Location:   req.Location,
		StreamPath: req.StreamPath,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.Created(w, "device created", result)
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	result, err := h.devices.List(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "devices loaded", result)
}

func (h *DeviceHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	result, err := h.devices.Get(r.Context(), claims.UserID, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "device loaded", result)
}

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	var req deviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.devices.Update(r.Context(), claims.UserID, chi.URLParam(r, "id"), service.DeviceInput{
		Name:       req.Name,
		DeviceID:   req.DeviceID,
		Type:       req.Type,
		Location:   req.Location,
		StreamPath: req.StreamPath,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "device updated", result)
}

func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	if err := h.devices.Delete(r.Context(), claims.UserID, chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "device deleted", nil)
}
