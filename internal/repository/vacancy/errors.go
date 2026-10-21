package vacancy

import "errors"

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrVacancyNotFound  = errors.New("vacancy not found")
)
