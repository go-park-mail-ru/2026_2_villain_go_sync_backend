package vacancy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	vacancyusecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/vacancy"
)

type VacancyUsecase interface {
	List(ctx context.Context) ([]models.Vacancy, error)
	Create(ctx context.Context, input vacancyusecase.CreateVacancyInput) (models.Vacancy, error)
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
}

type Handler struct {
	vacancy VacancyUsecase
}

func NewHandler(vacancy VacancyUsecase) *Handler {
	return &Handler{vacancy: vacancy}
}

func (h *Handler) ListVacancies(w http.ResponseWriter, r *http.Request) {
	vacancies, err := h.vacancy.List(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, vacancies)
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

	input := vacancyusecase.CreateVacancyInput{
		EmployerID:  userID,
		Title:       request.Title,
		Description: request.Description,
		SalaryFrom:  request.SalaryFrom,
		SalaryTo:    request.SalaryTo,
		Currency:    request.Currency,
		CategoryIDs: request.CategoryIDs,
	}

	createdVacancy, err := h.vacancy.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, vacancyusecase.ErrInvalidVacancy):
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, vacancyusecase.ErrEmployerProfileRequired):
			httputil.WriteError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, vacancyusecase.ErrNotEmployer):
			httputil.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	httputil.WriteOK(w, http.StatusCreated, createdVacancy)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid vacancy id")
		return
	}
	vacancy, err := h.vacancy.GetByID(r.Context(), id)
	if errors.Is(err, vacancyusecase.ErrVacancyNotFound) {
		httputil.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteOK(w, http.StatusOK , vacancy)

}
