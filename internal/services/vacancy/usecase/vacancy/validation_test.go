package vacancy

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
)

func TestValidateVacancyText(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		wantErr     bool
	}{
		{
			name:  "valid",
			title: "Go Developer",
		},
		{
			name:    "empty title",
			title:   "   ",
			wantErr: true,
		},
		{
			name:  "title exactly 100",
			title: strings.Repeat("Я", 100),
		},
		{
			name:    "title 101",
			title:   strings.Repeat("Я", 101),
			wantErr: true,
		},
		{
			name:        "description exactly 10000",
			title:       "Go Developer",
			description: strings.Repeat("Я", 10000),
		},
		{
			name:        "description 10001",
			title:       "Go Developer",
			description: strings.Repeat("Я", 10001),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVacancyText(tt.title, tt.description)

			if tt.wantErr {
				if !errors.Is(err, apperrors.ErrInvalidInput) {
					t.Fatalf("expected ErrInvalidInput, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
