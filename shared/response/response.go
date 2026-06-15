package response

import (
	"encoding/json"
	"net/http"
)

// Body is the standard JSON shape returned by every AgriPro service.
type Body struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// OK sends a successful response with optional data.
func OK(w http.ResponseWriter, message string, data any) {
	JSON(w, http.StatusOK, Body{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 response, typically after registration.
func Created(w http.ResponseWriter, message string, data any) {
	JSON(w, http.StatusCreated, Body{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends a failed response with the matching HTTP status.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, Body{
		Success: false,
		Message: message,
	})
}
