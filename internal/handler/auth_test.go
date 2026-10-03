package handler

import (
	"errors"
	"strings"
	"testing"
)

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"minimum length", "Abcdefg1", true},
		{"maximum length", "Aa1" + strings.Repeat("a", 69), true},
		{"empty", "", false},
		{"too short", "Aa1", false},
		{"too long", "Aa1" + strings.Repeat("a", 70), false},
		{"no uppercase", "password123", false},
		{"no lowercase", "PASSWORD123", false},
		{"no digit", "PasswordABC", false},
		{"forbidden character", "Password123 ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPassword(tt.password); got != tt.want {
				t.Errorf("isValidPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid", "test@example.com", true},
		{"maximum length", strings.Repeat("a", 242) + "@example.com", true},
		{"too long", strings.Repeat("a", 243) + "@example.com", false},
		{"empty", "", false},
		{"missing at", "test.example.com", false},
		{"missing local part", "@example.com", false},
		{"missing domain", "test@", false},
		{"contains space", "test user@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidEmail(tt.email); got != tt.want {
				t.Errorf("isValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{"seeker", true},
		{"employer", true},
		{"admin", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run("role="+tt.role, func(t *testing.T) {
			if got := isValidRole(tt.role); got != tt.want {
				t.Errorf("isValidRole(%q) = %v, want %v",
					tt.role, got, tt.want)
			}
		})
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{"valid", "test@example.com", "Password123", false},
		{"invalid email", "invalid", "Password123", true},
		{"invalid password", "test@example.com", "short", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCredentials(tt.email, tt.password)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidCredentials) {
					t.Errorf("error = %v, want ErrInvalidCredentials", err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
