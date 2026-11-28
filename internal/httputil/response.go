package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/requestid"
)

type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSON(context.Background(), w, status, payload, 0)
}

func writeJSON(ctx context.Context, w http.ResponseWriter, status int, payload any, limit int) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		logResponseError(ctx, "serialize HTTP response", err)
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if limit > 0 && len(encoded)+1 > limit {
		WriteError(w, http.StatusServiceUnavailable, "dependency unavailable")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(encoded, '\n')); err != nil {
		logResponseError(ctx, "write HTTP response", err)
	}
}

func logResponseError(ctx context.Context, operation string, err error) {
	slog.ErrorContext(ctx, operation, "request_id", requestid.FromContext(ctx), "error_type", fmt.Sprintf("%T", err))
}

func WriteOK(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, response{Data: data})
}

func WriteLimitedOK(ctx context.Context, w http.ResponseWriter, status int, data any, limit int) {
	writeJSON(ctx, w, status, response{Data: data}, limit)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, response{Error: msg})
}

type ErrorPolicy struct {
	DeadlineAsUnavailable bool
	InvalidInputStatus    int
	Message               string
	NotFoundMessage       string
	ForbiddenMessage      string
	ConflictMessage       string
}

func WriteAppError(w http.ResponseWriter, err error, msg string) {
	WriteAppErrorContext(context.Background(), w, err, ErrorPolicy{
		InvalidInputStatus: http.StatusBadRequest, Message: msg, NotFoundMessage: msg, ForbiddenMessage: msg, ConflictMessage: msg,
	})
}

func WriteAppErrorContext(ctx context.Context, w http.ResponseWriter, err error, policy ErrorPolicy) {
	var maxErr *http.MaxBytesError
	var bad BadRequest
	var validationErr apperrors.ValidationError
	switch {
	case errors.As(err, &bad):
		WriteError(w, http.StatusBadRequest, bad.Error())
	case errors.As(err, &maxErr):
		WriteError(w, http.StatusRequestEntityTooLarge, policy.Message)
	case errors.Is(err, apperrors.ErrDependencyUnavailable),
		policy.DeadlineAsUnavailable && errors.Is(err, context.DeadlineExceeded):
		WriteError(w, http.StatusServiceUnavailable, "dependency unavailable")
	case errors.As(err, &validationErr):
		WriteError(w, http.StatusUnprocessableEntity, validationErr.Error())
	case errors.Is(err, apperrors.ErrInvalidInput):
		WriteError(w, policy.InvalidInputStatus, policy.Message)
	case errors.Is(err, apperrors.ErrNotFound), errors.Is(err, apperrors.ErrNotOwner):
		WriteError(w, http.StatusNotFound, policy.NotFoundMessage)
	case errors.Is(err, apperrors.ErrFeatureDisabled):
		WriteError(w, http.StatusForbidden, "feature disabled")
	case errors.Is(err, apperrors.ErrForbidden):
		WriteError(w, http.StatusForbidden, policy.ForbiddenMessage)
	case errors.Is(err, apperrors.ErrConflict), errors.Is(err, apperrors.ErrRetryable):
		WriteError(w, http.StatusConflict, policy.ConflictMessage)
	case errors.Is(err, apperrors.ErrNoContent):
		w.WriteHeader(http.StatusNoContent)
	default:
		logResponseError(ctx, "handle HTTP request", err)
		WriteError(w, http.StatusInternalServerError, "internal error")
	}
}
