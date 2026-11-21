package apperrors

import (
	"errors"
	"fmt"
)

var (
	ErrEmailTaken       = errors.New("email already taken")
	ErrNotFound         = errors.New("entity not found")
	ErrForbidden        = errors.New("forbidden")
	ErrNotOwner         = fmt.Errorf("resource ownership: %w", ErrForbidden)
	ErrNoContent        = errors.New("No content")
	ErrConflict         = errors.New("conflict")
	ErrCategoryNotFound = errors.New("category not found")
	ErrInvalidInput     = errors.New("invalid input")
)

type ValidationError string

func (err ValidationError) Error() string {
	return string(err)
}

func (err ValidationError) Unwrap() error {
	return ErrInvalidInput
}
