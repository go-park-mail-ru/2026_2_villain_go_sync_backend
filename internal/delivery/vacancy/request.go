package vacancy

import (
	"encoding/json"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

type CreateVacancyRequest struct {
	Title       sanitizer.SafeString `json:"title"`
	Description sanitizer.SafeString `json:"description"`
	SalaryFrom  *float64             `json:"salary_from"`
	SalaryTo    *float64             `json:"salary_to"`
	Currency    sanitizer.SafeString `json:"currency"`
	CategoryIDs []int64              `json:"category_ids"`
}

type UpdateVacancyRequest struct {
	Title       *sanitizer.SafeString `json:"title"`
	Description *sanitizer.SafeString `json:"description"`
	SalaryFrom  json.RawMessage       `json:"salary_from"`
	SalaryTo    json.RawMessage       `json:"salary_to"`
	Currency    *sanitizer.SafeString `json:"currency"`
	CategoryIDs *[]int64              `json:"category_ids"`
}
