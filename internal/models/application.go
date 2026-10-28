package models

import "time"

type Application struct {
	ID             int64     `json:"id"`
	VacancyID      int64     `json:"vacancy_id"`
	VacancyVersion int       `json:"vacancy_version"`
	ResumeID       int64     `json:"resume_id"`
	CoverLetter    *string   `json:"cover_letter"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
