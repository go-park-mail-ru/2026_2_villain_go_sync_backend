package resume

import (
	"encoding/json"
	"errors"
	"strings"

	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/structured"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

type UpsertRequest struct {
	structured.CareerDetails
	Title          sanitizer.SafeString  `json:"title"`
	Description    sanitizer.SafeString  `json:"description"`
	Experience     sanitizer.SafeString  `json:"experience"`
	Education      sanitizer.SafeString  `json:"education"`
	TargetPosition sanitizer.SafeString  `json:"target_position"`
	City           *sanitizer.SafeString `json:"city"`
	EmploymentType *sanitizer.SafeString `json:"employment_type"`
	WorkFormat     *sanitizer.SafeString `json:"work_format"`
	WorkSchedule   *sanitizer.SafeString `json:"work_schedule"`
	SalaryFrom     json.RawMessage       `json:"salary_from"`
	PortfolioURL   *sanitizer.SafeString `json:"portfolio_url"`
	Languages      *sanitizer.SafeString `json:"languages"`
}

func (r UpsertRequest) resume() (models.Resume, error) {
	if strings.TrimSpace(r.Title.String()) == "" {
		return models.Resume{}, errors.New("title is required")
	}
	targetPosition := strings.TrimSpace(r.TargetPosition.String())
	if targetPosition == "" {
		return models.Resume{}, errors.New("target_position is required")
	}

	switch r.EmploymentType.String() {
	case "", "full", "part":
	default:
		return models.Resume{}, errors.New("employment_type must be full or part")
	}
	switch r.WorkFormat.String() {
	case "", "office", "hybrid", "remote":
	default:
		return models.Resume{}, errors.New("work_format must be office, hybrid or remote")
	}
	switch r.WorkSchedule.String() {
	case "", "5_2", "2_2", "flexible", "shift":
	default:
		return models.Resume{}, errors.New("invalid work_schedule")
	}

	var salary *float64
	if r.SalaryFrom != nil {
		if err := json.Unmarshal(r.SalaryFrom, &salary); err != nil {
			return models.Resume{}, errors.New("salary_from must be a number or null")
		}
		if salary != nil && (*salary < 0 || *salary > 9999999999.99) {
			return models.Resume{}, errors.New("salary_from is out of range")
		}
	}
	details, err := r.CareerDetails.Model(false)
	if err != nil {
		return models.Resume{}, err
	}
	return models.Resume{
		Title:  r.Title.String(),
		Career: details,
		Content: models.ResumeContent{
			Description:  r.Description.String(),
			Experience:   r.Experience.String(),
			Education:    r.Education.String(),
			PortfolioURL: sanitizer.Ptr(r.PortfolioURL),
			Languages:    sanitizer.Ptr(r.Languages),
		},
		Preferences: models.ResumePreferences{
			TargetPosition: &targetPosition,
			City:           sanitizer.Ptr(r.City),
			EmploymentType: sanitizer.Ptr(r.EmploymentType),
			WorkFormat:     sanitizer.Ptr(r.WorkFormat),
			WorkSchedule:   sanitizer.Ptr(r.WorkSchedule),
			SalaryFrom:     salary,
		},
	}, nil
}

func (r UpsertRequest) update(id int) (dto.UpdateInput, error) {
	resume, err := r.resume()
	if err != nil {
		return dto.UpdateInput{}, err
	}
	resume.ID = id
	return dto.UpdateInput{
		Resume:          resume,
		ClearSalaryFrom: r.SalaryFrom != nil && resume.Preferences.SalaryFrom == nil,
	}, nil
}
