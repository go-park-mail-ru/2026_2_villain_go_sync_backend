package handler

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
)

func (h *Handler) ListVacancies(w http.ResponseWriter, r *http.Request) {
	vacancies, err := h.Vacancies.List()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	httputil.WriteOK(w, http.StatusOK, vacancies)
}
