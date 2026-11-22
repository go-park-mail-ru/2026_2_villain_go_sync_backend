package vacancy

import (
	"encoding/json"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

type CreateVacancyRequest struct {
	Title            sanitizer.SafeString  `json:"title"`
	Description      sanitizer.SafeString  `json:"description"`
	Responsibilities sanitizer.SafeString  `json:"responsibilities"`
	Requirements     sanitizer.SafeString  `json:"requirements"`
	SalaryFrom       *float64              `json:"salary_from"`
	SalaryTo         *float64              `json:"salary_to"`
	Currency         sanitizer.SafeString  `json:"currency"`
	Offer            *sanitizer.SafeString `json:"offer"`
	ExperienceLevel  *sanitizer.SafeString `json:"experience_level"`
	EmploymentType   *sanitizer.SafeString `json:"employment_type"`
	WorkFormat       *sanitizer.SafeString `json:"work_format"`
	WorkSchedule     *sanitizer.SafeString `json:"work_schedule"`
	WorkHours        *sanitizer.SafeString `json:"work_hours"`
	City             *sanitizer.SafeString `json:"city"`
	OfficeAddress    *sanitizer.SafeString `json:"office_address"`
	CategoryIDs      []int64               `json:"category_ids"`
}

type UpdateVacancyRequest struct {
	Title            *sanitizer.SafeString `json:"title"`
	Description      *sanitizer.SafeString `json:"description"`
	Responsibilities *sanitizer.SafeString `json:"responsibilities"`
	Requirements     *sanitizer.SafeString `json:"requirements"`
	SalaryFrom       json.RawMessage       `json:"salary_from"`
	SalaryTo         json.RawMessage       `json:"salary_to"`
	Currency         *sanitizer.SafeString `json:"currency"`
	Offer            *sanitizer.SafeString `json:"offer"`
	ExperienceLevel  *sanitizer.SafeString `json:"experience_level"`
	EmploymentType   *sanitizer.SafeString `json:"employment_type"`
	WorkFormat       *sanitizer.SafeString `json:"work_format"`
	WorkSchedule     *sanitizer.SafeString `json:"work_schedule"`
	WorkHours        *sanitizer.SafeString `json:"work_hours"`
	City             *sanitizer.SafeString `json:"city"`
	OfficeAddress    *sanitizer.SafeString `json:"office_address"`
	CategoryIDs      *[]int64              `json:"category_ids"`
}
