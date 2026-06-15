package handler

import (
	"net/http"

	"github.com/19parwiz/agripro-core/services/farm/internal/service"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	sharedmiddleware "github.com/19parwiz/agripro-core/shared/middleware"
	"github.com/19parwiz/agripro-core/shared/response"
	"github.com/go-chi/chi/v5"
)

type PlantHandler struct {
	plants *service.PlantService
}

func NewPlantHandler(plants *service.PlantService) *PlantHandler {
	return &PlantHandler{plants: plants}
}

type plantRequest struct {
	Name         string `json:"name"`
	Variety      string `json:"variety"`
	PlantingDate string `json:"planting_date"`
}

func (h *PlantHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	var req plantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	plantingDate, err := parseDate(req.PlantingDate)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.plants.Create(r.Context(), claims.UserID, service.PlantInput{
		Name:         req.Name,
		Variety:      req.Variety,
		PlantingDate: plantingDate,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.Created(w, "plant created", result)
}

func (h *PlantHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	result, err := h.plants.List(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "plants loaded", result)
}

func (h *PlantHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	result, err := h.plants.Get(r.Context(), claims.UserID, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "plant loaded", result)
}

func (h *PlantHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	var req plantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	plantingDate, err := parseDate(req.PlantingDate)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.plants.Update(r.Context(), claims.UserID, chi.URLParam(r, "id"), service.PlantInput{
		Name:         req.Name,
		Variety:      req.Variety,
		PlantingDate: plantingDate,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "plant updated", result)
}

func (h *PlantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, apperrors.ErrUnauthorized)
		return
	}

	if err := h.plants.Delete(r.Context(), claims.UserID, chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "plant deleted", nil)
}
