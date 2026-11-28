package delivery

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/requestid"
)

const adminPrefix = "/api/onlineconf/admin"
const maxBodySize = 1 << 20

type endpoint func(*http.Request, int64) (any, error)

type handler struct {
	usecases usecase.Usecases
}

func New(u usecase.Usecases, tokens *auth.TokenManager, snapshotSecret string, health func(context.Context) error) http.Handler {
	h := handler{usecases: u}
	mux := http.NewServeMux()
	register := func(method, path string, status int, run endpoint) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actorID, ok := auth.UserIDFromContext(r.Context())
			if !ok || actorID <= 0 {
				httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
			data, err := run(r, actorID)
			if err != nil {
				writeError(w, r, err)
				return
			}
			if status == http.StatusNoContent {
				w.WriteHeader(status)
				return
			}
			httputil.WriteOK(w, status, data)
		})
		mux.Handle(method+" "+adminPrefix+path, middleware.Auth(tokens, middleware.CSRF(next)))
	}
	h.registerAdmin(register)

	secretHash := sha256.Sum256([]byte(snapshotSecret))
	mux.HandleFunc("GET /internal/onlineconf/snapshot", func(w http.ResponseWriter, r *http.Request) {
		scheme, secret, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		hash := sha256.Sum256([]byte(secret))
		if snapshotSecret == "" || !ok || !strings.EqualFold(scheme, "Bearer") || subtle.ConstantTimeCompare(hash[:], secretHash[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		snapshot, err := u.Snapshots.Get(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}
		httputil.WriteOK(w, http.StatusOK, snapshotView(snapshot))
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := health(ctx); err != nil {
			httputil.WriteError(w, http.StatusServiceUnavailable, "dependency unavailable")
			return
		}
		httputil.WriteOK(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{"ok"})
	})
	return middleware.RequestID(middleware.AccessLog(mux, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(ctx))
	})))
}

type badRequest string

func (e badRequest) Error() string { return string(e) }

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var bad badRequest
	var validation apperrors.ValidationError
	switch {
	case errors.As(err, &bad):
		httputil.WriteError(w, http.StatusBadRequest, bad.Error())
	case errors.Is(err, repository.ErrUnavailable), errors.Is(err, usecase.ErrProductUsersUnavailable),
		errors.Is(err, context.DeadlineExceeded):
		httputil.WriteError(w, http.StatusServiceUnavailable, "dependency unavailable")
	case errors.Is(err, apperrors.ErrForbidden):
		httputil.WriteError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, apperrors.ErrNotFound):
		httputil.WriteError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, apperrors.ErrConflict), errors.Is(err, repository.ErrRetryable):
		httputil.WriteError(w, http.StatusConflict, "conflict")
	case errors.As(err, &validation):
		httputil.WriteError(w, http.StatusUnprocessableEntity, validation.Error())
	case errors.Is(err, apperrors.ErrInvalidInput):
		httputil.WriteError(w, http.StatusUnprocessableEntity, "invalid input")
	default:
		slog.ErrorContext(r.Context(), "onlineconf request failed", "request_id", requestid.FromContext(r.Context()), "error_type", fmt.Sprintf("%T", err))
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}

func decode[T any](r *http.Request) (T, error) {
	var input *T
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || input == nil {
		var zero T
		return zero, badRequest("invalid JSON body")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		var zero T
		return zero, badRequest("invalid JSON body")
	}
	return *input, nil
}

func positiveID(raw, name string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, badRequest("invalid " + name)
	}
	return id, nil
}

func pathID(r *http.Request, name string) (int64, error) {
	return positiveID(r.PathValue(name), name)
}

func editID(r *http.Request) (int64, error) {
	if r.Method == http.MethodPost {
		return 0, nil
	}
	return pathID(r, "id")
}

func pageRequest(r *http.Request) (repository.PageRequest, error) {
	q := r.URL.Query()
	var page repository.PageRequest
	if q.Has("limit") {
		limit, err := positiveID(q.Get("limit"), "limit")
		if err != nil || limit > 100 {
			return page, badRequest("limit must be from 1 to 100")
		}
		page.Limit = int(limit)
	}
	if q.Has("cursor") {
		cursor, err := positiveID(q.Get("cursor"), "cursor")
		if err != nil {
			return page, err
		}
		page.Cursor = &cursor
	}
	return page, nil
}

func byID[T, R any](param string, get func(context.Context, int64, int64) (T, error), view func(T) R) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		id, err := pathID(r, param)
		if err != nil {
			return nil, err
		}
		result, err := get(r.Context(), actor, id)
		return view(result), err
	}
}

func remove(param string, run func(context.Context, int64, int64) error) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		id, err := pathID(r, param)
		if err != nil {
			return nil, err
		}
		return nil, run(r.Context(), actor, id)
	}
}

func list[T, R any](get func(context.Context, int64, repository.PageRequest) (repository.Page[T], error), view func(T) R) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		page, err := pageRequest(r)
		if err != nil {
			return nil, err
		}
		result, err := get(r.Context(), actor, page)
		return pageView(result, view), err
	}
}

func listByID[T, R any](param string,
	get func(context.Context, int64, int64, repository.PageRequest) (repository.Page[T], error), view func(T) R,
) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		id, err := pathID(r, param)
		if err != nil {
			return nil, err
		}
		page, err := pageRequest(r)
		if err != nil {
			return nil, err
		}
		result, err := get(r.Context(), actor, id, page)
		return pageView(result, view), err
	}
}
