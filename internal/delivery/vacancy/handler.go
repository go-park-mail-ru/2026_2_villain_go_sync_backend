package vacancy

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type VacancyUsecase interface {
	List(ctx context.Context) ([]models.Vacancy, error)
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
