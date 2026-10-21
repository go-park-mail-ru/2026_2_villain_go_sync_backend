package vacancy

import "errors"

var (
	ErrInvalidVacancy          = errors.New("invalid vacancy")
	ErrNotEmployer             = errors.New("only employers can create vacancies")
	ErrEmployerProfileRequired = errors.New("employer profile is required")
	ErrVacancyNotFound         = errors.New("vacancy not found")
)
