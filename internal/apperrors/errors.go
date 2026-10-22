package apperrors

import "errors"

var (
	ErrEmailTaken = errors.New("email already taken")
	ErrNotFound   = errors.New("entity not found")
	ErrForbidden  = errors.New("forbidden")
	ErrNoContent  = errors.New("No content")
)
