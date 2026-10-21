package vacancy

import "errors"

var (
	ErrInvalidVacancy          = errors.New("invalid vacancy")
	ErrNotEmployer             = errors.New("only employers can manage vacancies")
	ErrEmployerProfileRequired = errors.New("employer profile is required")
	ErrVacancyNotFound         = errors.New("vacancy not found")
	ErrNotVacancyOwner         = errors.New("vacancy does not belong to this employer")
	ErrVacancyVersionConflict  = errors.New("vacancy version conflict")
)
