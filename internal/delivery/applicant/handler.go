package applicant

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ApplicantUsecase interface {
	List(ctx context.Context, userID int, role models.Role, vacancyID int) ([]models.Applicant, error)
}

type Handler struct {
	applicants ApplicantUsecase
}

func NewHandler(applicants ApplicantUsecase) *Handler {
	return &Handler{applicants: applicants}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	vacancyID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	applicants, err := h.applicants.List(r.Context(), int(userID), role, vacancyID)
	if errors.Is(err, apperrors.ErrForbidden) {
		httputil.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, applicants)
}
