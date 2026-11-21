package models

import (
	"fmt"
	"time"
)

type Vacancy struct {
	ID         int64
	Version    int
	Title      string
	ApplyUntil *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ChangedAt  time.Time

	Company      VacancyCompany
	Content      VacancyContent
	Requirements VacancyRequirements
	Salary       VacancySalary
	Conditions   VacancyConditions
	Location     VacancyLocation
}

type VacancyCompany struct {
	EmployerID int64
	Name       string
	LogoURL    *string
}

type VacancyContent struct {
	Description      string
	Responsibilities string
}

type VacancyRequirements struct {
	Description     string
	ExperienceLevel *ExperienceLevel
}

type VacancySalary struct {
	SalaryFrom *float64
	SalaryTo   *float64
	Currency   Currency
}

type VacancyConditions struct {
	Offer          *string
	EmploymentType *EmploymentType
	WorkFormat     *WorkFormat
	WorkSchedule   *WorkSchedule
	WorkHours      *string
}

type VacancyLocation struct {
	City          *string
	OfficeAddress *string
}

func (v *Vacancy) SetLogoURL(logoKey string) {
	if logoKey == "" {
		v.Company.LogoURL = nil
		return
	}

	url := fmt.Sprintf("/api/employers/%d/logo", v.Company.EmployerID)
	v.Company.LogoURL = &url
}
