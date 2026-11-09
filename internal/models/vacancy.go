package models

import (
	"fmt"
	"time"
)

type Vacancy struct {
	ID          int64     `json:"id"`
	Version     int       `json:"version"`
	EmployerID  int64     `json:"employer_id"`
	CompanyName string    `json:"company_name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	SalaryFrom  *float64  `json:"salary_from"`
	SalaryTo    *float64  `json:"salary_to"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	LogoURL     *string   `json:"logo_url"`
}

func (v *Vacancy) SetLogoURL(logoKey string) {
	if logoKey == "" {
		v.LogoURL = nil
		return
	}

	url := fmt.Sprintf("/api/employers/%d/logo", v.EmployerID)
	v.LogoURL = &url
}
