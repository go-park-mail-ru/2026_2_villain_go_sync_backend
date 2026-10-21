package apperrors

import "errors"

var (
	ErrEmailTaken             = errors.New("email already taken")
	ErrNotFound               = errors.New("user not found")
	ErrVacancyNotFound        = errors.New("vacancy not found")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrNotVacancyOwner        = errors.New("vacancy does not belong to this employer")
	ErrVacancyVersionConflict = errors.New("vacancy version conflict")
)
