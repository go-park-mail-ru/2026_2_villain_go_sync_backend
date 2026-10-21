package vacancy

import "errors"

var (
	ErrCategoryNotFound       = errors.New("category not found")
	ErrVacancyNotFound        = errors.New("vacancy not found")
	ErrNotVacancyOwner        = errors.New("vacancy does not belong to this employer")
	ErrVacancyVersionConflict = errors.New("vacancy version conflict")
)
