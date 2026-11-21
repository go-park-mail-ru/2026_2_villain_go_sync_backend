package profile_test

import (
	"bytes"
	"encoding/json"
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
				Profile:      models.Profile{UserID: 12},
				PersonalInfo: models.SeekerPersonalInfo{FirstName: "Ivan", AvatarKey: tt.key},
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
				Profile: models.Profile{UserID: 11},
				Company: models.EmployerCompanyInfo{CompanyName: "Test Company", LogoKey: tt.key},
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

func TestProfileResponseJSON(t *testing.T) {
	seeker := profile.NewSeekerResponse(models.SeekerProfile{
		Profile:      models.Profile{UserID: 12},
		PersonalInfo: models.SeekerPersonalInfo{FirstName: "Ivan", AvatarKey: ptr("private/avatar")},
		Contacts:     models.SeekerContacts{Phone: "123"},
		Preferences:  models.SeekerPreferences{WorkFormat: (*models.WorkFormat)(ptr("remote"))},
		Career:       models.CareerDetails{Skills: []models.NamedItem{{Name: "Go"}}},
	})
	employer := profile.NewEmployerResponse(models.EmployerProfile{
		Profile:  models.Profile{UserID: 11},
		Company:  models.EmployerCompanyInfo{CompanyName: "Company", LogoKey: ptr("private/logo")},
		Contacts: models.EmployerContacts{Website: "example.com"},
		Location: models.EmployerLocation{MainCity: ptr("Moscow")},
		Details: models.CompanyDetails{
			Technologies: []models.NamedItem{{Name: "Go"}},
			Culture:      []models.CultureItem{{ShortDescription: ptr("Open")}},
		},
	})

	for _, tc := range []struct {
		name     string
		response any
		fields   map[string]string
	}{
		{"seeker", seeker, map[string]string{
			"user_id": "12", "first_name": `"Ivan"`, "phone": `"123"`,
			"avatar_url": `"/api/seekers/12/avatar"`, "work_format": `"remote"`, "skills": `[{"name":"Go"}]`,
		}},
		{"employer", employer, map[string]string{
			"user_id": "11", "company_name": `"Company"`, "website": `"example.com"`,
			"logo_url": `"/api/employers/11/logo"`, "main_city": `"Moscow"`,
			"technologies": `[{"name":"Go"}]`, "culture": `[{"title":null,"short_description":"Open","description":null}]`,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.response)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]json.RawMessage
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			for key, want := range tc.fields {
				if value := string(got[key]); value != want {
					t.Errorf("%s = %s, want %s", key, value, want)
				}
			}
			if bytes.Contains(raw, []byte("private/")) {
				t.Error("storage key leaked into JSON")
			}
		})
	}
}
