package vacancy

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	applicationdto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

//go:generate mockgen -source=handler.go -destination=mock_handler.go -package=vacancy
type VacancyUsecase interface {
	ListVacancies(ctx context.Context) ([]models.Vacancy, error)
	Create(ctx context.Context, input dto.CreateVacancyInput) (models.Vacancy, error)
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
	ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error)
	Update(ctx context.Context, input dto.UpdateVacancyInput) (models.Vacancy, error)
	Delete(ctx context.Context, vacancyID, employerID int64) error
	ListApplicants(ctx context.Context, userID int64, role models.Role, vacancyID int64) ([]models.Applicant, error)
	ListEmployerApplicants(ctx context.Context, userID int64, role models.Role) ([]models.Applicant, error)
}

type Handler struct {
	vacancy VacancyUsecase
}

func NewHandler(vacancy VacancyUsecase) *Handler {
	return &Handler{vacancy: vacancy}
}

func (h *Handler) ListVacancies(w http.ResponseWriter, r *http.Request) {
	vacancies, err := h.vacancy.ListVacancies(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, dto.NewVacancyListResponse(vacancies))
}

func (h *Handler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	var request CreateVacancyRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input := dto.CreateVacancyInput{
		EmployerID:       userID,
		Title:            request.Title.String(),
		Description:      request.Description.String(),
		Responsibilities: request.Responsibilities.String(),
		Requirements:     request.Requirements.String(),
		SalaryFrom:       request.SalaryFrom,
		SalaryTo:         request.SalaryTo,
		Currency:         request.Currency.String(),
		Offer:            sanitizer.Ptr(request.Offer),
		ExperienceLevel:  sanitizer.Ptr(request.ExperienceLevel),
		EmploymentType:   sanitizer.Ptr(request.EmploymentType),
		WorkFormat:       sanitizer.Ptr(request.WorkFormat),
		WorkSchedule:     sanitizer.Ptr(request.WorkSchedule),
		WorkHours:        sanitizer.Ptr(request.WorkHours),
		City:             sanitizer.Ptr(request.City),
		OfficeAddress:    sanitizer.Ptr(request.OfficeAddress),
		CategoryIDs:      request.CategoryIDs,
	}

	createdVacancy, err := h.vacancy.Create(r.Context(), input)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to create vacancy")
		return
	}
	httputil.WriteOK(w, http.StatusCreated, dto.NewVacancyResponse(createdVacancy))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}
	vacancy, err := h.vacancy.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteAppError(w, err, "vacancy not found")
		return
	}

	httputil.WriteOK(w, http.StatusOK, dto.NewVacancyResponse(vacancy))

}

func (h *Handler) ListVacanciesByEmployer(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vacancies, err := h.vacancy.ListByEmployer(r.Context(), userID)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to list vacancies")
		return
	}

	httputil.WriteOK(w, http.StatusOK, dto.NewVacancyListResponse(vacancies))
}

func parseOptionalSalary(raw json.RawMessage) (dto.OptionalSalary, error) {
	if len(raw) == 0 {
		return dto.OptionalSalary{Provided: false}, nil
	}
	if string(raw) == "null" {
		return dto.OptionalSalary{Provided: true, Value: nil}, nil
	}

	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		return dto.OptionalSalary{}, err
	}

	return dto.OptionalSalary{Provided: true, Value: &value}, nil
}

func (h *Handler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	var request UpdateVacancyRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vacancyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || vacancyID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	salaryFrom, err := parseOptionalSalary(request.SalaryFrom)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	salaryTo, err := parseOptionalSalary(request.SalaryTo)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := dto.UpdateVacancyInput{
		VacancyID:        vacancyID,
		EmployerID:       userID,
		Title:            sanitizer.Ptr(request.Title),
		Description:      sanitizer.Ptr(request.Description),
		Responsibilities: sanitizer.Ptr(request.Responsibilities),
		Requirements:     sanitizer.Ptr(request.Requirements),
		SalaryFrom:       salaryFrom,
		SalaryTo:         salaryTo,
		Currency:         sanitizer.Ptr(request.Currency),
		Offer:            sanitizer.Ptr(request.Offer),
		ExperienceLevel:  sanitizer.Ptr(request.ExperienceLevel),
		EmploymentType:   sanitizer.Ptr(request.EmploymentType),
		WorkFormat:       sanitizer.Ptr(request.WorkFormat),
		WorkSchedule:     sanitizer.Ptr(request.WorkSchedule),
		WorkHours:        sanitizer.Ptr(request.WorkHours),
		City:             sanitizer.Ptr(request.City),
		OfficeAddress:    sanitizer.Ptr(request.OfficeAddress),
		CategoryIDs:      request.CategoryIDs,
	}

	updatedVacancy, err := h.vacancy.Update(r.Context(), input)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to update vacancy")
		return
	}
	httputil.WriteOK(w, http.StatusOK, dto.NewVacancyResponse(updatedVacancy))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vacancyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || vacancyID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}

	err = h.vacancy.Delete(r.Context(), vacancyID, userID)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to delete vacancy")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListApplicants(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := auth.RoleFromContext(r.Context())

	vacancyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	applicants, err := h.vacancy.ListApplicants(r.Context(), userID, role, vacancyID)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to list applicants")
		return
	}

	httputil.WriteOK(w, http.StatusOK, applicationdto.NewApplicantListResponse(applicants))
}

func (h *Handler) ListEmployerApplicants(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, ok := auth.RoleFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	applicants, err := h.vacancy.ListEmployerApplicants(r.Context(), userID, role)
	if err != nil {
		httputil.WriteAppError(w, err, "failed to list applicants")
		return
	}

	httputil.WriteOK(w, http.StatusOK, applicationdto.NewApplicantListResponse(applicants))
}
