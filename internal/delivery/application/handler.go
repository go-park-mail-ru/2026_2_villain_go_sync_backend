package application

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ApplicationUsecase interface {
	Create(ctx context.Context, input dto.CreateApplicationInput) (models.Application, error)
}

type Handler struct {
	applications ApplicationUsecase
}

func NewHandler(applications ApplicationUsecase) *Handler {
	return &Handler{applications: applications}
}

type CreateApplicationRequest struct {
	ResumeID    int     `json:"resume_id"`
	CoverLetter *string `json:"cover_letter"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	seekerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vacancyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || vacancyID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	var request CreateApplicationRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	created, err := h.applications.Create(r.Context(), dto.CreateApplicationInput{
		SeekerID:    seekerID,
		VacancyID:   vacancyID,
		ResumeID:    request.ResumeID,
		CoverLetter: request.CoverLetter,
	})

	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrInvalidInput):
			httputil.WriteError(w, http.StatusBadRequest, "invalid application data")
		case errors.Is(err, apperrors.ErrForbidden):
			httputil.WriteError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, apperrors.ErrNotFound):
			httputil.WriteError(w, http.StatusNotFound, "vacancy or resume not found")
		case errors.Is(err, apperrors.ErrConflict):
			httputil.WriteError(w, http.StatusConflict, "already applied with this resume")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	httputil.WriteOK(w, http.StatusCreated, created)
}
