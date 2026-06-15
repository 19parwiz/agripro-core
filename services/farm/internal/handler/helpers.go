package handler

import (
	"encoding/json"
	"net/http"
	"time"

	apperrors "github.com/19parwiz/agripro-core/shared/errors"
	"github.com/19parwiz/agripro-core/shared/response"
)

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

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, apperrors.ErrBadRequest
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, apperrors.New(http.StatusBadRequest, "planting_date must be YYYY-MM-DD")
	}

	return parsed, nil
}
