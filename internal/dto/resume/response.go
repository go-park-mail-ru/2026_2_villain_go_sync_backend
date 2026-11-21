package resume

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/structured"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ResumeResponse struct {
	ID             int       `json:"id"`
	SeekerID       int       `json:"seeker_id"`
	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Description    string    `json:"description"`
	Experience     string    `json:"experience"`
	Education      string    `json:"education"`
	PortfolioURL   *string   `json:"portfolio_url"`
	Languages      *string   `json:"languages"`
	TargetPosition *string   `json:"target_position"`
	City           *string   `json:"city"`
	EmploymentType *string   `json:"employment_type"`
	WorkFormat     *string   `json:"work_format"`
	WorkSchedule   *string   `json:"work_schedule"`
	SalaryFrom     *float64  `json:"salary_from"`
	structured.CareerDetailsResponse
}

func NewResumeResponse(r models.Resume) ResumeResponse {
	return ResumeResponse{
		ID:                    r.ID,
		SeekerID:              r.SeekerID,
		Title:                 r.Title,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		Description:           r.Content.Description,
		Experience:            r.Content.Experience,
		Education:             r.Content.Education,
		PortfolioURL:          r.Content.PortfolioURL,
		Languages:             r.Content.Languages,
		TargetPosition:        r.Preferences.TargetPosition,
		City:                  r.Preferences.City,
		EmploymentType:        r.Preferences.EmploymentType,
		WorkFormat:            r.Preferences.WorkFormat,
		WorkSchedule:          r.Preferences.WorkSchedule,
		SalaryFrom:            r.Preferences.SalaryFrom,
		CareerDetailsResponse: structured.NewCareerDetailsResponse(r.Career),
	}
}

func NewResumeListResponse(resumes []models.Resume) []ResumeResponse {
	result := make([]ResumeResponse, len(resumes))
	for i, r := range resumes {
		result[i] = NewResumeResponse(r)
	}
	return result
}
