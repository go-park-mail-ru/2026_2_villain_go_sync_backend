package profile_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestNewSeekerResponse(t *testing.T) {
	tests := []struct {
		name    string
		key     *string
		wantURL string
	}{
		{name: "no avatar"},
		{name: "empty key", key: ptr("")},
		{name: "with avatar", key: ptr("avatars/12/a.png"), wantURL: "/api/seekers/12/avatar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.NewSeekerResponse(models.SeekerProfile{
				Profile:   models.Profile{UserID: 12},
				FirstName: "Ivan",
				AvatarKey: tt.key,
			})
			if got.FirstName != "Ivan" {
				t.Errorf("lost first name: %q", got.FirstName)
			}
			if tt.wantURL == "" {
				if got.AvatarURL != nil {
					t.Errorf("wanted nil URL, got %q", *got.AvatarURL)
				}
			} else if got.AvatarURL == nil || *got.AvatarURL != tt.wantURL {
				t.Errorf("want URL %q, got %v", tt.wantURL, got.AvatarURL)
			}
		})
	}
}

func TestNewEmployerResponse(t *testing.T) {
	tests := []struct {
		name    string
		key     *string
		wantURL string
	}{
		{name: "no logo"},
		{name: "empty key", key: ptr("")},
		{name: "with logo", key: ptr("logos/11/a.png"), wantURL: "/api/employers/11/logo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.NewEmployerResponse(models.EmployerProfile{
				Profile:     models.Profile{UserID: 11},
				CompanyName: "Test Company",
				LogoKey:     tt.key,
			})
			if got.CompanyName != "Test Company" {
				t.Errorf("lost company name: %q", got.CompanyName)
			}
			if tt.wantURL == "" {
				if got.LogoURL != nil {
					t.Errorf("wanted nil URL, got %q", *got.LogoURL)
				}
			} else if got.LogoURL == nil || *got.LogoURL != tt.wantURL {
				t.Errorf("want URL %q, got %v", tt.wantURL, got.LogoURL)
			}
		})
	}
}

func ptr(s string) *string { return &s }
