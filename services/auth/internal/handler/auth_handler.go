package handler

import (
	"encoding/json"
	"net/http"

	"github.com/19parwiz/agripro-core/services/auth/internal/service"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	"github.com/19parwiz/agripro-core/shared/response"
)

// AuthHandler exposes auth endpoints over HTTP.
type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.authService.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.Created(w, "account created, please verify your email", result)
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	result, err := h.authService.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "login successful", result)
}

// VerifyEmail handles POST /api/auth/verify-email.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		writeError(w, err)
		return
	}

	response.OK(w, "email verified successfully", nil)
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return apperrors.ErrBadRequest
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return apperrors.New(http.StatusBadRequest, "invalid request body")
	}

	return nil
}

func writeError(w http.ResponseWriter, err error) {
	response.Error(w, apperrors.HTTPStatus(err), apperrors.Message(err))
}
