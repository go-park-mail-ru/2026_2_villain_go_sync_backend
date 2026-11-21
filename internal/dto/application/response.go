package application

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ApplicationResponse struct {
	SeekerID       int64     `json:"seeker_id"`
	ID             int64     `json:"id"`
	VacancyID      int64     `json:"vacancy_id"`
	VacancyVersion int       `json:"vacancy_version"`
	ResumeID       int64     `json:"resume_id"`
	CoverLetter    *string   `json:"cover_letter"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func NewApplicationResponse(a models.Application) ApplicationResponse {
	return ApplicationResponse{
		SeekerID:       a.SeekerID,
		ID:             a.ID,
		VacancyID:      a.VacancyID,
		VacancyVersion: a.VacancyVersion,
		ResumeID:       a.ResumeID,
		CoverLetter:    a.CoverLetter,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

type ApplicationStatusResponse struct {
	Status          string    `json:"status"`
	ReasonCode      *string   `json:"reason_code"`
	StatusComment   *string   `json:"status_comment"`
	StatusChangedAt time.Time `json:"status_changed_at"`
}

func NewApplicationStatusResponse(a models.ApplicationStatus) ApplicationStatusResponse {
	return ApplicationStatusResponse{
		Status:          a.Value,
		ReasonCode:      a.ReasonCode,
		StatusComment:   a.StatusComment,
		StatusChangedAt: a.StatusChangedAt,
	}
}

type ApplicationStatusUpdateResponse struct {
	ApplicationID int64 `json:"application_id"`
	ApplicationStatusResponse
}

func NewApplicationStatusUpdateResponse(a models.ApplicationStatusUpdate) ApplicationStatusUpdateResponse {
	return ApplicationStatusUpdateResponse{
		ApplicationID:             a.ApplicationID,
		ApplicationStatusResponse: NewApplicationStatusResponse(a.ApplicationStatus),
	}
}

type SeekerApplicationResponse struct {
	ID             int64     `json:"id"`
	SeekerID       int64     `json:"seeker_id"`
	CoverLetter    *string   `json:"cover_letter"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	VacancyID      int64     `json:"vacancy_id"`
	VacancyVersion int       `json:"vacancy_version"`
	VacancyTitle   string    `json:"vacancy_title"`
	CompanyName    string    `json:"company_name"`
	ResumeID       *int64    `json:"resume_id"`
	ResumeTitle    string    `json:"resume_title"`
	ApplicationStatusResponse
}

func NewSeekerApplicationResponse(a models.SeekerApplication) SeekerApplicationResponse {
	return SeekerApplicationResponse{
		ID:                        a.ID,
		SeekerID:                  a.SeekerID,
		CoverLetter:               a.CoverLetter,
		CreatedAt:                 a.CreatedAt,
		UpdatedAt:                 a.UpdatedAt,
		VacancyID:                 a.Vacancy.VacancyID,
		VacancyVersion:            a.Vacancy.VacancyVersion,
		VacancyTitle:              a.Vacancy.VacancyTitle,
		CompanyName:               a.Vacancy.CompanyName,
		ResumeID:                  a.Resume.ResumeID,
		ResumeTitle:               a.Resume.ResumeTitle,
		ApplicationStatusResponse: NewApplicationStatusResponse(a.Status),
	}
}

func NewSeekerApplicationListResponse(applications []models.SeekerApplication) []SeekerApplicationResponse {
	if applications == nil {
		return nil
	}
	items := make([]SeekerApplicationResponse, len(applications))
	for i, application := range applications {
		items[i] = NewSeekerApplicationResponse(application)
	}
	return items
}
