package vacancy

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

const maxSalary = 9999999999.99

func validateSalary(from, to *float64) error {
	if from != nil && *from < 0 {
		return fmt.Errorf("%w: salary_from must be non-negative", apperrors.ErrInvalidInput)
	}

	if to != nil && *to < 0 {
		return fmt.Errorf("%w: salary_to must be non-negative", apperrors.ErrInvalidInput)
	}

	if from != nil && *from > maxSalary {
		return fmt.Errorf("%w: salary_from is too large", apperrors.ErrInvalidInput)
	}

	if to != nil && *to > maxSalary {
		return fmt.Errorf("%w: salary_to is too large", apperrors.ErrInvalidInput)
	}

	if from != nil && to != nil && *from > *to {
		return fmt.Errorf("%w: salary_from must not exceed salary_to", apperrors.ErrInvalidInput)
	}

	if from != nil && math.Round(*from*100)/100 != *from {
		return fmt.Errorf("%w: salary_from must have at most 2 decimal places", apperrors.ErrInvalidInput)
	}

	if to != nil && math.Round(*to*100)/100 != *to {
		return fmt.Errorf("%w: salary_to must have at most 2 decimal places", apperrors.ErrInvalidInput)
	}
	return nil
}

const (
	maxVacancyTitleLength       = 100
	maxVacancyDescriptionLength = 10000
)

func validateVacancyText(title, description string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", apperrors.ErrInvalidInput)
	}

	if utf8.RuneCountInString(title) > maxVacancyTitleLength {
		return fmt.Errorf("%w: title must not exceed 100 characters", apperrors.ErrInvalidInput)
	}

	if utf8.RuneCountInString(description) > maxVacancyDescriptionLength {
		return fmt.Errorf("%w: description must not exceed 10000 characters", apperrors.ErrInvalidInput)
	}

	return nil
}

func validateCategories(categoryIDs []int64) error {
	if len(categoryIDs) > 5 {
		return fmt.Errorf("%w: too many categories", apperrors.ErrInvalidInput)
	}

	seen := make(map[int64]bool)
	for _, categoryID := range categoryIDs {
		if categoryID <= 0 {
			return fmt.Errorf("%w: invalid category id", apperrors.ErrInvalidInput)
		}
		if seen[categoryID] {
			return fmt.Errorf("%w: duplicate category id", apperrors.ErrInvalidInput)
		}
		seen[categoryID] = true
	}
	return nil
}

func validateVacancyDetails(responsibilities, requirements string) error {
	if strings.TrimSpace(responsibilities) == "" {
		return fmt.Errorf("%w: responsibilities are required", apperrors.ErrInvalidInput)
	}
	if strings.TrimSpace(requirements) == "" {
		return fmt.Errorf("%w: requirements are required", apperrors.ErrInvalidInput)
	}
	return nil
}

func validateVacancyOptions(vacancy models.Vacancy) error {
	if vacancy.Requirements.ExperienceLevel != nil {
		switch *vacancy.Requirements.ExperienceLevel {
		case "", "no_experience", "1_3", "3_6", "6_plus":
		default:
			return fmt.Errorf("%w: invalid experience_level", apperrors.ErrInvalidInput)
		}
	}
	if vacancy.Conditions.EmploymentType != nil {
		switch *vacancy.Conditions.EmploymentType {
		case "", "full", "part":
		default:
			return fmt.Errorf("%w: invalid employment_type", apperrors.ErrInvalidInput)
		}
	}
	if vacancy.Conditions.WorkFormat != nil {
		switch *vacancy.Conditions.WorkFormat {
		case "", "office", "hybrid", "remote":
		default:
			return fmt.Errorf("%w: invalid work_format", apperrors.ErrInvalidInput)
		}
	}
	if vacancy.Conditions.WorkSchedule != nil {
		switch *vacancy.Conditions.WorkSchedule {
		case "", "5_2", "2_2", "flexible", "shift":
		default:
			return fmt.Errorf("%w: invalid work_schedule", apperrors.ErrInvalidInput)
		}
	}
	return nil
}
