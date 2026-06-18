package handler

import (
	"encoding/json"
	"net/http"
	"strings"
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

func parseHistoryRange(r *http.Request, defaultRange time.Duration) (time.Time, time.Time, error) {
	query := r.URL.Query()
	now := time.Now().UTC()

	if preset := strings.TrimSpace(query.Get("range")); preset != "" {
		duration, err := parsePresetRange(preset)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return now.Add(-duration), now, nil
	}

	from, err := parseTimeQuery(query.Get("from"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if from.IsZero() {
		from = now.Add(-defaultRange)
	}

	to, err := parseTimeQuery(query.Get("to"))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if to.IsZero() {
		to = now
	}

	if !to.After(from) {
		return time.Time{}, time.Time{}, apperrors.ErrBadRequest
	}

	return from, to, nil
}

func parsePresetRange(value string) (time.Duration, error) {
	switch strings.ToLower(value) {
	case "24h":
		return 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	case "6mo", "180d":
		return 180 * 24 * time.Hour, nil
	default:
		return 0, apperrors.New(http.StatusBadRequest, "range must be 24h, 7d, 30d, or 6mo")
	}
}

func parseTimeQuery(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, apperrors.New(http.StatusBadRequest, "time must be RFC3339 or YYYY-MM-DD")
	}

	return parsed.UTC(), nil
}

func requiredQuery(r *http.Request, key string) (string, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return "", apperrors.New(http.StatusBadRequest, key+" is required")
	}
	return value, nil
}
