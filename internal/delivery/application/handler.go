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
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

//go:generate mockgen -source=handler.go -destination=mock_handler.go -package=application
type ApplicationUsecase interface {
	UpdateStatus(ctx context.Context, role models.Role, input dto.UpdateStatusInput) (models.ApplicationStatusUpdate, error)
	ListBySeeker(ctx context.Context, seekerID int64, role models.Role, page int, limit int) ([]models.SeekerApplication, int64, error)
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

type UpdateStatusRequest struct {
	Status     string                `json:"status"`
	ReasonCode *string               `json:"reason_code"`
	Comment    *sanitizer.SafeString `json:"comment"`
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	employerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid application id")
		return
	}

	var request UpdateStatusRequest
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		var maxBytes *http.MaxBytesError
		if errors.As(err, &maxBytes) {
			httputil.WriteError(w, http.StatusRequestEntityTooLarge, "request too large")
		} else {
			httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		}
		return
	}

	updated, err := h.applications.UpdateStatus(r.Context(), role, dto.UpdateStatusInput{
		ApplicationID: id,
		EmployerID:    employerID,
		Status:        request.Status,
		ReasonCode:    request.ReasonCode,
		Comment:       sanitizer.Ptr(request.Comment),
	})
	if errors.Is(err, apperrors.ErrInvalidInput) {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		httputil.WriteAppError(w, err, "failed to update application status")
		return
	}
	httputil.WriteOK(w, http.StatusOK, updated)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	page, limit, ok := httputil.ReadPagination(w, r)
	if !ok {
		return
	}
	applications, total, err := h.applications.ListBySeeker(r.Context(), userID, role, page, limit)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to list applications")
		return
	}
	httputil.WritePage(w, applications, page, limit, total)
}
