package profile

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	profiledto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ProfileUsecase interface {
	GetSeekerByID(ctx context.Context, seekerID int) (models.SeekerProfile, error)
	UpdateSeeker(ctx context.Context, userID int, role models.Role, profile models.SeekerProfile) (models.SeekerProfile, error)
	GetEmployerByID(ctx context.Context, employerID int) (models.EmployerProfile, error)
	UpdateEmployer(ctx context.Context, userID int, role models.Role, profile models.EmployerProfile) (models.EmployerProfile, error)
	GetSeekerAvatar(ctx context.Context, userID int) (io.ReadCloser, string, error)
	UploadSeekerAvatar(ctx context.Context, userID int, role models.Role, body io.ReadSeeker, contentType string) error
	DeleteSeekerAvatar(ctx context.Context, userID int, role models.Role) error
	GetEmployerLogo(ctx context.Context, userID int) (io.ReadCloser, string, error)
	UploadEmployerLogo(ctx context.Context, userID int, role models.Role, body io.ReadSeeker, contentType string) error
	DeleteEmployerLogo(ctx context.Context, userID int, role models.Role) error
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
	if err != nil {
		httputil.WriteAppError(w, err, "seeker not found")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profiledto.NewSeekerResponse(profile))
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
			Description: req.Description.String(),
		},
		FirstName: req.FirstName.String(),
		LastName:  req.LastName.String(),
		Phone:     req.Phone.String(),
	})
	if err != nil {
		httputil.WriteAppError(w, err, "failed to update seeker profile")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profiledto.NewSeekerResponse(profile))
}

func (h *Handler) GetPublicEmployer(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	employerID, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	profile, err := h.profiles.GetEmployerByID(r.Context(), employerID)
	if err != nil {
		httputil.WriteAppError(w, err, "employer not found")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profiledto.NewEmployerResponse(profile))
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
			Description: req.Description.String(),
		},
		CompanyName: req.CompanyName.String(),
		Website:     req.Website.String(),
	})
	if err != nil {
		httputil.WriteAppError(w, err, "failed to update employer profile")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profiledto.NewEmployerResponse(profile))
}

func (h *Handler) GetSeekerAvatar(w http.ResponseWriter, r *http.Request) {
	seekerID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	body, contentType, err := h.profiles.GetSeekerAvatar(r.Context(), seekerID)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to get seeker avatar")
		return
	}

	defer body.Close()
	w.Header().Set("Content-Type", contentType)
	_, err = io.Copy(w, body)
	if err != nil {
		slog.Error("failed to hand over seeker avatar",
			"user_id", seekerID,
			"error", err,
		)
	}
}

func (h *Handler) UploadSeekerAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	file, contentType, ok := openProfileImage(w, r, "avatar")
	if !ok {
		return
	}
	defer file.Close()

	err := h.profiles.UploadSeekerAvatar(r.Context(), int(userID), role, file, contentType)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to upload seeker avatar")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteSeekerAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())
	err := h.profiles.DeleteSeekerAvatar(r.Context(), int(userID), role)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to delete seeker avatar")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetEmployerLogo(w http.ResponseWriter, r *http.Request) {
	employerID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	body, contentType, err := h.profiles.GetEmployerLogo(r.Context(), employerID)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to get employer logo")
		return
	}

	defer body.Close()
	w.Header().Set("Content-Type", contentType)
	_, err = io.Copy(w, body)
	if err != nil {
		slog.Error("failed to hand over employer logo",
			"user_id", employerID,
			"error", err,
		)
	}
}

func (h *Handler) UploadEmployerLogo(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	file, contentType, ok := openProfileImage(w, r, "logo")
	if !ok {
		return
	}
	defer file.Close()

	err := h.profiles.UploadEmployerLogo(r.Context(), int(userID), role, file, contentType)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to upload employer logo")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteEmployerLogo(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())
	err := h.profiles.DeleteEmployerLogo(r.Context(), int(userID), role)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to delete employer logo")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
