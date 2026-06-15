package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError carries an HTTP status and a message handlers can return to the client.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates an error with an HTTP status and a client-safe message.
func New(code int, message string) error {
	return &AppError{Code: code, Message: message}
}

// Wrap adds context to an existing error while keeping the HTTP status.
func Wrap(err error, code int, message string) error {
	return &AppError{Code: code, Message: message, Err: err}
}

// HTTPStatus returns the status code from an AppError, or 500 for unknown errors.
func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// Message returns the client-safe message from an AppError.
func Message(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "something went wrong"
}

// Common errors used across services.
var (
	ErrNotFound          = New(http.StatusNotFound, "resource not found")
	ErrUnauthorized      = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden         = New(http.StatusForbidden, "forbidden")
	ErrBadRequest        = New(http.StatusBadRequest, "invalid request")
	ErrConflict          = New(http.StatusConflict, "resource already exists")
	ErrInternal          = New(http.StatusInternalServerError, "internal server error")
	ErrInvalidCredentials = New(http.StatusUnauthorized, "invalid email or password")
	ErrEmailNotVerified  = New(http.StatusForbidden, "email not verified")
)
