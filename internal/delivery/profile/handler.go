package profile

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

type ProfileUsecase interface {
	GetSeekerByID(ctx context.Context, seekerID int) (models.SeekerProfile, error)
	UpdateSeeker(ctx context.Context, userID int, role models.Role, profile models.SeekerProfile) (models.SeekerProfile, error)
	GetEmployerByID(ctx context.Context, employerID int) (models.EmployerProfile, error)
	UpdateEmployer(ctx context.Context, userID int, role models.Role, profile models.EmployerProfile) (models.EmployerProfile, error)
}

type Handler struct {
	profiles ProfileUsecase
}

func NewHandler(profiles ProfileUsecase) *Handler {
	return &Handler{profiles: profiles}
}

func (h *Handler) GetSeekerByID(w http.ResponseWriter, r *http.Request) {
	seekerID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	profile, err := h.profiles.GetSeekerByID(r.Context(), seekerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "seeker not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profile)
}

func (h *Handler) UpdateSeeker(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	var req UpdateSeekerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := req.IsValid(); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	profile, err := h.profiles.UpdateSeeker(r.Context(), int(userID), role, models.SeekerProfile{
		Profile: models.Profile{
			UserID:      int(userID),
			Description: req.Description,
		},
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	})
	if errors.Is(err, apperrors.ErrForbidden) {
		httputil.WriteError(w, http.StatusForbidden, "only seekers can update this profile")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profile)
}

func (h *Handler) GetPublicEmployer(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	employerID, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	profile, err := h.profiles.GetEmployerByID(r.Context(), employerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "employer not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profile)
}

func (h *Handler) UpdateEmployer(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	var req UpdateEmployerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := req.IsValid(); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	profile, err := h.profiles.UpdateEmployer(r.Context(), int(userID), role, models.EmployerProfile{
		Profile: models.Profile{
			UserID:      int(userID),
			Description: req.Description,
		},
		CompanyName: req.CompanyName,
		Website:     req.Website,
	})
	if errors.Is(err, apperrors.ErrForbidden) {
		httputil.WriteError(w, http.StatusForbidden, "only employers can update this profile")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profile)
}
