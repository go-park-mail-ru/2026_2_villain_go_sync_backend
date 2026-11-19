package models

import (
	"fmt"
	"time"
)

type Vacancy struct {
	ID               int64     `json:"id"`
	Version          int       `json:"version"`
	EmployerID       int64     `json:"employer_id"`
	CompanyName      string    `json:"company_name"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Responsibilities string    `json:"responsibilities"`
	Requirements     string    `json:"requirements"`
	SalaryFrom       *float64  `json:"salary_from"`
	SalaryTo         *float64  `json:"salary_to"`
	Currency         string    `json:"currency"`
	Offer            *string   `json:"offer"`
	ExperienceLevel  *string   `json:"experience_level"`
	EmploymentType   *string   `json:"employment_type"`
	WorkFormat       *string   `json:"work_format"`
	WorkSchedule     *string   `json:"work_schedule"`
	WorkHours        *string   `json:"work_hours"`
	City             *string   `json:"city"`
	OfficeAddress    *string   `json:"office_address"`
	ApplyUntil       *string   `json:"apply_until"`
	UpdatedAt        time.Time `json:"updated_at"`
	ChangedAt        time.Time `json:"changed_at"`
	CreatedAt        time.Time `json:"created_at"`
	LogoURL          *string   `json:"logo_url"`
}

func (v *Vacancy) SetLogoURL(logoKey string) {
	if logoKey == "" {
		v.LogoURL = nil
		return
	}

	url := fmt.Sprintf("/api/employers/%d/logo", v.EmployerID)
	v.LogoURL = &url
}
