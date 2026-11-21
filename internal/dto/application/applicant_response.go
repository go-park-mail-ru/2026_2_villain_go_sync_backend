package application

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ApplicantResponse struct {
	ApplicationID   int64      `json:"application_id"`
	CoverLetter     *string    `json:"cover_letter"`
	AppliedAt       time.Time  `json:"applied_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	SeekerID        int64      `json:"seeker_id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	AvatarURL       *string    `json:"avatar_url"`
	VacancyID       int64      `json:"vacancy_id"`
	VacancyTitle    string     `json:"vacancy_title"`
	VacancyVersion  int        `json:"vacancy_version"`
	ResumeID        *int64     `json:"resume_id"`
	ResumeTitle     string     `json:"resume_title"`
	Status          string     `json:"status"`
	StatusHistoryID *int64     `json:"status_history_id"`
	StatusComment   *string    `json:"status_comment"`
	StatusChangedAt *time.Time `json:"status_changed_at"`
	StatusChangedBy *int64     `json:"status_changed_by"`
	ReasonCode      *string    `json:"reason_code"`
}

func NewApplicantResponse(a models.Applicant) ApplicantResponse {
	return ApplicantResponse{
		ApplicationID:   a.ApplicationID,
		CoverLetter:     a.CoverLetter,
		AppliedAt:       a.AppliedAt,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
		SeekerID:        a.Seeker.SeekerID,
		FirstName:       a.Seeker.FirstName,
		LastName:        a.Seeker.LastName,
		AvatarURL:       a.Seeker.AvatarURL,
		VacancyID:       a.Vacancy.VacancyID,
		VacancyTitle:    a.Vacancy.VacancyTitle,
		VacancyVersion:  a.Vacancy.VacancyVersion,
		ResumeID:        a.Resume.ResumeID,
		ResumeTitle:     a.Resume.ResumeTitle,
		Status:          string(a.Status.Value),
		StatusHistoryID: a.Status.StatusHistoryID,
		StatusComment:   a.Status.StatusComment,
		StatusChangedAt: a.Status.StatusChangedAt,
		StatusChangedBy: a.Status.StatusChangedBy,
		ReasonCode:      (*string)(a.Status.ReasonCode),
	}
}

func NewApplicantListResponse(applicants []models.Applicant) []ApplicantResponse {
	if applicants == nil {
		return nil
	}
	items := make([]ApplicantResponse, len(applicants))
	for i, applicant := range applicants {
		items[i] = NewApplicantResponse(applicant)
	}
	return items
}
