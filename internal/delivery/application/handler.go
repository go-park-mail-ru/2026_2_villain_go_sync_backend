package application

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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
		httputil.WriteAppError(w, err, "failed to create application")
		return
	}

	httputil.WriteOK(w, http.StatusCreated, created)
}
