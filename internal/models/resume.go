package models

import "time"

type Resume struct {
	CareerDetails
	ID              int       `json:"id"`
	SeekerID        int       `json:"seeker_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Experience      string    `json:"experience"`
	Education       string    `json:"education"`
	TargetPosition  *string   `json:"target_position"`
	City            *string   `json:"city"`
	EmploymentType  *string   `json:"employment_type"`
	WorkFormat      *string   `json:"work_format"`
	WorkSchedule    *string   `json:"work_schedule"`
	SalaryFrom      *float64  `json:"salary_from"`
	PortfolioURL    *string   `json:"portfolio_url"`
	Languages       *string   `json:"languages"`
	ClearSalaryFrom bool      `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
