package models

import "time"

type Application struct {
	SeekerID       int64
	ID             int64
	VacancyID      int64
	VacancyVersion int
	ResumeID       int64
	CoverLetter    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ApplicationStatus struct {
	Value           ApplicationStatusValue
	ReasonCode      *RejectionReasonCode
	StatusComment   *string
	StatusChangedAt time.Time
}

type ApplicationStatusUpdate struct {
	ApplicationID int64
	ApplicationStatus
}

type SeekerApplication struct {
	ID          int64
	SeekerID    int64
	CoverLetter *string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Vacancy ApplicationVacancyInfo
	Resume  ApplicationResumeInfo
	Status  ApplicationStatus
}

type ApplicationVacancyInfo struct {
	VacancyID      int64
	VacancyVersion int
	VacancyTitle   string
	CompanyName    string
}

type ApplicationResumeInfo struct {
	ResumeID    *int64
	ResumeTitle string
}
