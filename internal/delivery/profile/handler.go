package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	GetSeekerAvatar(ctx context.Context, userID int) (io.ReadCloser, string, error)
	UploadSeekerAvatar(ctx context.Context, userID int, role models.Role, body io.ReadSeeker, contentType string) error
	DeleteSeekerAvatar(ctx context.Context, userID int, role models.Role) error
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
	if err != nil {
		httputil.WriteAppError(w, err, "employer not found")
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
			Description: req.Description.String(),
		},
		CompanyName: req.CompanyName.String(),
		Website:     req.Website.String(),
	})
	if err != nil {
		httputil.WriteAppError(w, err, "failed to update employer profile")
		return
	}

	httputil.WriteOK(w, http.StatusOK, profile)
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

const maxProfileImageSize int64 = 10 << 20

func (h *Handler) UploadSeekerAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, maxProfileImageSize+(1<<20))
	file, header, err := r.FormFile("avatar")
	if err != nil {
		httputil.WriteAppError(
			w,
			fmt.Errorf("%w: %w", apperrors.ErrInvalidInput, err),
			"invalid avatar upload",
		)
		return
	}
	defer file.Close()

	if header.Size > maxProfileImageSize {
		httputil.WriteError(w, http.StatusRequestEntityTooLarge, "avatar is too large")
		return
	}
	if header.Size == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "avatar is empty")
		return
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil &&
		!errors.Is(err, io.EOF) &&
		!errors.Is(err, io.ErrUnexpectedEOF) {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read file")
		return
	}
	if n == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "avatar is empty")
		return
	}

	contentType := http.DetectContentType(buf[:n])
	switch contentType {
	case "image/png", "image/jpeg", "image/webp":
	default:
		httputil.WriteError(w, http.StatusUnsupportedMediaType, "unsupported avatar format")
		return
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reset avatar file")
		return
	}
	err = h.profiles.UploadSeekerAvatar(r.Context(), int(userID), role, file, contentType)
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
