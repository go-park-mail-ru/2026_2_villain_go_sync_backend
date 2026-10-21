package auth

import (
	"regexp"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

const (
	maxEmailLength    = 254
	minPasswordLength = 8
	maxPasswordLength = 72
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var (
	reUpper = regexp.MustCompile(`[A-Z]`)
	reLower = regexp.MustCompile(`[a-z]`)
	reDigit = regexp.MustCompile(`[0-9]`)
	reChars = regexp.MustCompile("^[A-Za-z0-9!@#$%^&*()_+\\-=\\[\\]{};':\"\\\\|,.<>/?`~]+$")
)

var allowedRoles = map[string]bool{
	"employer": true,
	"seeker":   true,
}

func validateCredentials(email, pass string) error {
	if !isValidEmail(email) || !isValidPassword(pass) {
		return ErrInvalidRequest
	}
	return nil
}

func isValidEmail(email string) bool {
	if len(email) > maxEmailLength {
		return false
	}
	return emailRegex.MatchString(email)
}

func isValidPassword(password string) bool {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return false
	}
	if !reUpper.MatchString(password) {
		return false
	}
	if !reLower.MatchString(password) {
		return false
	}
	if !reDigit.MatchString(password) {
		return false
	}
	if !reChars.MatchString(password) {
		return false
	}
	return true
}

func isValidRole(role models.Role) bool {
	switch role {
	case models.RoleSeeker, models.RoleEmployer:
		return true
	default:
		return false
	}
}
