package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	if message == "" {
		message = http.StatusText(status)
	}

	writeJSON(w, status, errorResponse{Error: message})
}

type errInvalidQuery string

func (e errInvalidQuery) Error() string {
	return string(e)
}

func IsInvalidQuery(err error) bool {
	var target errInvalidQuery
	return errors.As(err, &target)
}
