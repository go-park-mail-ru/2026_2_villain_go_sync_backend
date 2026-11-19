package resume

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

//go:generate mockgen -source=handler.go -destination=mock_handler.go -package=resume
type ResumeUsecase interface {
	Create(ctx context.Context, userID int, role models.Role, resume models.Resume) (models.Resume, error)
	GetByID(ctx context.Context, id int, userID int, role models.Role) (models.Resume, error)
	ListBySeekerID(ctx context.Context, seekerID int) ([]models.Resume, error)
	Update(ctx context.Context, userID int, role models.Role, resume models.Resume) (models.Resume, error)
	Delete(ctx context.Context, userID int, role models.Role, id int) error
}

type Handler struct {
	resumes ResumeUsecase
}

func NewHandler(resumes ResumeUsecase) *Handler {
	return &Handler{resumes: resumes}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	var req UpserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := req.IsValid(); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	res, err := h.resumes.Create(r.Context(), int(userID), role, models.Resume{
		Title:       req.Title.String(),
		Description: req.Description.String(),
		Experience:  req.Experience.String(),
		Education:   req.Education.String(),
	})
	if err != nil {
		httputil.WriteAppError(w, err, "failed to create resume")
		return
	}

	httputil.WriteOK(w, http.StatusCreated, res)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	res, err := h.resumes.GetByID(r.Context(), id, int(userID), role)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrForbidden):
			httputil.WriteError(w, http.StatusForbidden, "access denied")
		case errors.Is(err, apperrors.ErrNotFound):
			httputil.WriteError(w, http.StatusNotFound, "resume not found")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	httputil.WriteOK(w, http.StatusOK, res)
}

func (h *Handler) ListBySeekerID(w http.ResponseWriter, r *http.Request) {
	seekerIDStr := r.URL.Query().Get("seeker_id")
	if seekerIDStr == "" {
		httputil.WriteError(w, http.StatusBadRequest, "seeker_id query param required")
		return
	}
	seekerID, err := strconv.Atoi(seekerIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid seeker_id")
		return
	}

	resumes, err := h.resumes.ListBySeekerID(r.Context(), seekerID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, resumes)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := req.IsValid(); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	res, err := h.resumes.Update(r.Context(), int(userID), role, models.Resume{
		ID:          id,
		Title:       req.Title.String(),
		Description: req.Description.String(),
		Experience:  req.Experience.String(),
		Education:   req.Education.String(),
	})
	if err != nil {
		httputil.WriteAppError(w, err, "failed to update resume")
		return
	}

	httputil.WriteOK(w, http.StatusOK, res)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.resumes.Delete(r.Context(), int(userID), role, id)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to delete resume")
		return
	}

	httputil.WriteOK(w, http.StatusNoContent, nil)
}
