package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) ListVacancies(w http.ResponseWriter, r *http.Request) {
	vacancies, err := h.Vacancies.List()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
