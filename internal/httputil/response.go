package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
)

type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteOK(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, response{Data: data})
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, response{Error: msg})
}

func WriteAppError(w http.ResponseWriter, err error, msg string) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidInput):
		WriteError(w, http.StatusBadRequest, msg)
	case errors.Is(err, apperrors.ErrForbidden):
		WriteError(w, http.StatusForbidden, msg)
	case errors.Is(err, apperrors.ErrNotFound):
		WriteError(w, http.StatusNotFound, msg)
	case errors.Is(err, apperrors.ErrConflict):
		WriteError(w, http.StatusConflict, msg)
	case errors.Is(err, apperrors.ErrNoContent):
		w.WriteHeader(http.StatusNoContent)
	default:
		WriteError(w, http.StatusInternalServerError, "internal error")
	}
}
