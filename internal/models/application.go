package models

import "time"

type Application struct {
	SeekerID       int64     `json:"seeker_id"`
	ID             int64     `json:"id"`
	VacancyID      int64     `json:"vacancy_id"`
	VacancyVersion int       `json:"vacancy_version"`
	ResumeID       int64     `json:"resume_id"`
	CoverLetter    *string   `json:"cover_letter"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ApplicationStatus struct {
	Status          string    `json:"status"`
	ReasonCode      *string   `json:"reason_code"`
	StatusComment   *string   `json:"status_comment"`
	StatusChangedAt time.Time `json:"status_changed_at"`
}

type ApplicationStatusUpdate struct {
	ApplicationID int64 `json:"application_id"`
	ApplicationStatus
}

type SeekerApplication struct {
	ID             int64     `json:"id"`
	SeekerID       int64     `json:"seeker_id"`
	VacancyID      int64     `json:"vacancy_id"`
	VacancyVersion int       `json:"vacancy_version"`
	CoverLetter    *string   `json:"cover_letter"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ResumeID       *int64    `json:"resume_id"`
	VacancyTitle   string    `json:"vacancy_title"`
	CompanyName    string    `json:"company_name"`
	ResumeTitle    string    `json:"resume_title"`
	ApplicationStatus
}
