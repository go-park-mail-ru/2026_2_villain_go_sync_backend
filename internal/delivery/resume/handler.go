package resume

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ResumeUsecase interface {
	Create(ctx context.Context, userID int, role models.Role, resume models.Resume) (models.Resume, error)
	GetByID(ctx context.Context, id int) (models.Resume, error)
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
		Title:       req.Title,
		Description: req.Description,
		Experience:  req.Experience,
		Education:   req.Education,
	})
	if err != nil {
		httputil.WriteAppError(w, err, "failed to create resume")
		return
	}

	httputil.WriteOK(w, http.StatusCreated, res)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	res, err := h.resumes.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteAppError(w, err, "resume not found")
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
		Title:       req.Title,
		Description: req.Description,
		Experience:  req.Experience,
		Education:   req.Education,
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
