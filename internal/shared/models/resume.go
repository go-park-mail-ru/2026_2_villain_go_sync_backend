package models

import "time"

type Resume struct {
	ID        int
	SeekerID  int
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time

	Content     ResumeContent
	Preferences ResumePreferences
	Career      CareerDetails
}

type ResumeContent struct {
	Description  string
	Experience   string
	Education    string
	PortfolioURL *string
	Languages    *string
}

type ResumePreferences struct {
	TargetPosition *string
	City           *string
	EmploymentType *EmploymentType
	WorkFormat     *WorkFormat
	WorkSchedule   *WorkSchedule
	SalaryFrom     *float64
}
