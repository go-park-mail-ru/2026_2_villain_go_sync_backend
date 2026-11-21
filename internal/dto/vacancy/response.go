package vacancy

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type VacancyResponse struct {
	ID               int64     `json:"id"`
	Version          int       `json:"version"`
	Title            string    `json:"title"`
	ApplyUntil       *string   `json:"apply_until"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ChangedAt        time.Time `json:"changed_at"`
	EmployerID       int64     `json:"employer_id"`
	CompanyName      string    `json:"company_name"`
	LogoURL          *string   `json:"logo_url"`
	Description      string    `json:"description"`
	Responsibilities string    `json:"responsibilities"`
	Requirements     string    `json:"requirements"`
	ExperienceLevel  *string   `json:"experience_level"`
	SalaryFrom       *float64  `json:"salary_from"`
	SalaryTo         *float64  `json:"salary_to"`
	Currency         string    `json:"currency"`
	Offer            *string   `json:"offer"`
	EmploymentType   *string   `json:"employment_type"`
	WorkFormat       *string   `json:"work_format"`
	WorkSchedule     *string   `json:"work_schedule"`
	WorkHours        *string   `json:"work_hours"`
	City             *string   `json:"city"`
	OfficeAddress    *string   `json:"office_address"`
}

func NewVacancyResponse(v models.Vacancy) VacancyResponse {
	return VacancyResponse{
		ID:               v.ID,
		Version:          v.Version,
		Title:            v.Title,
		ApplyUntil:       v.ApplyUntil,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
		ChangedAt:        v.ChangedAt,
		EmployerID:       v.Company.EmployerID,
		CompanyName:      v.Company.Name,
		LogoURL:          v.Company.LogoURL,
		Description:      v.Content.Description,
		Responsibilities: v.Content.Responsibilities,
		Requirements:     v.Requirements.Description,
		ExperienceLevel:  (*string)(v.Requirements.ExperienceLevel),
		SalaryFrom:       v.Salary.SalaryFrom,
		SalaryTo:         v.Salary.SalaryTo,
		Currency:         string(v.Salary.Currency),
		Offer:            v.Conditions.Offer,
		EmploymentType:   (*string)(v.Conditions.EmploymentType),
		WorkFormat:       (*string)(v.Conditions.WorkFormat),
		WorkSchedule:     (*string)(v.Conditions.WorkSchedule),
		WorkHours:        v.Conditions.WorkHours,
		City:             v.Location.City,
		OfficeAddress:    v.Location.OfficeAddress,
	}
}

func NewVacancyListResponse(vacancies []models.Vacancy) []VacancyResponse {
	if vacancies == nil {
		return nil
	}
	items := make([]VacancyResponse, len(vacancies))
	for i, vacancy := range vacancies {
		items[i] = NewVacancyResponse(vacancy)
	}
	return items
}
