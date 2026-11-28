package delivery

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

const adminPrefix = "/api/onlineconf/admin"
const (
	maxBodySize            = 1 << 20
	maxConcurrentSnapshots = 2
	healthTimeout          = time.Second
	requestTimeout         = 10 * time.Second
)

var adminErrorPolicy = httputil.ErrorPolicy{
	DeadlineAsUnavailable: true,
	InvalidInputStatus:    http.StatusUnprocessableEntity,
	Message:               "invalid input",
	ForbiddenMessage:      "forbidden",
	NotFoundMessage:       "resource not found",
	ConflictMessage:       "conflict",
}

type endpoint func(*http.Request, int64) (any, error)

type handler struct {
	usecases Services
}

func New(u Services, tokens *auth.TokenManager, snapshotSecret string, health func(context.Context) error) http.Handler {
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
				httputil.WriteAppErrorContext(r.Context(), w, err, adminErrorPolicy)
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

	snapshotSlots := make(chan struct{}, maxConcurrentSnapshots)
	secretHash := sha256.Sum256([]byte(snapshotSecret))
	mux.HandleFunc("GET /internal/onlineconf/snapshot", func(w http.ResponseWriter, r *http.Request) {
		scheme, secret, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		hash := sha256.Sum256([]byte(secret))
		if snapshotSecret == "" || !ok || !strings.EqualFold(scheme, "Bearer") || subtle.ConstantTimeCompare(hash[:], secretHash[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		select {
		case snapshotSlots <- struct{}{}:
			defer func() { <-snapshotSlots }()
		default:
			httputil.WriteError(w, http.StatusServiceUnavailable, "dependency unavailable")
			return
		}
		snapshot, err := u.Snapshots.Get(r.Context())
		if err != nil {
			httputil.WriteAppErrorContext(r.Context(), w, err, adminErrorPolicy)
			return
		}
		httputil.WriteLimitedOK(r.Context(), w, http.StatusOK, snapshotView(snapshot), onlineconf.MaxSnapshotBytes)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
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
		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(ctx))
	})))
}

func editID(r *http.Request) (int64, error) {
	if r.Method == http.MethodPost {
		return 0, nil
	}
	return httputil.PathID(r, "id")
}

func pageRequest(r *http.Request) (models.PageRequest, error) {
	q, err := httputil.ScalarQuery(r, "cursor", "limit")
	if err != nil {
		return models.PageRequest{}, err
	}
	return parsePage(q)
}

func parsePage(q url.Values) (models.PageRequest, error) {
	var page models.PageRequest
	if q.Has("limit") {
		limit, err := httputil.PositiveID(q.Get("limit"), "limit")
		if err != nil || limit > models.MaxPageSize {
			return page, httputil.BadRequest("limit must be from 1 to " + strconv.Itoa(models.MaxPageSize))
		}
		page.Limit = int(limit)
	}
	if q.Has("cursor") {
		cursor, err := httputil.PositiveID(q.Get("cursor"), "cursor")
		if err != nil {
			return page, err
		}
		page.Cursor = &cursor
	}
	return page, nil
}

func byID[T, R any](param string, get func(context.Context, int64, int64) (T, error), view func(T) R) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		id, err := httputil.PathID(r, param)
		if err != nil {
			return nil, err
		}
		result, err := get(r.Context(), actor, id)
		return view(result), err
	}
}

func remove(param string, run func(context.Context, int64, int64) error) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		id, err := httputil.PathID(r, param)
		if err != nil {
			return nil, err
		}
		return nil, run(r.Context(), actor, id)
	}
}

func list[T, R any](get func(context.Context, int64, models.PageRequest) (models.Page[T], error), view func(T) R) endpoint {
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
	get func(context.Context, int64, int64, models.PageRequest) (models.Page[T], error), view func(T) R,
) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		id, err := httputil.PathID(r, param)
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
