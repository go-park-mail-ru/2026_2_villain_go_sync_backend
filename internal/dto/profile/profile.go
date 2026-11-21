package profile

import (
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/structured"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ProfileResponse struct {
	UserID      int       `json:"user_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newProfileResponse(p models.Profile) ProfileResponse {
	return ProfileResponse{
		UserID:      p.UserID,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

type SeekerResponse struct {
	ProfileResponse
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	City              *string `json:"city"`
	Phone             string  `json:"phone"`
	Messenger         *string `json:"messenger"`
	ContactEmail      *string `json:"contact_email"`
	PreferredPosition *string `json:"preferred_position"`
	PreferredCities   *string `json:"preferred_cities"`
	WorkFormat        *string `json:"work_format"`
	structured.CareerDetailsResponse
	AvatarURL *string `json:"avatar_url"`
}

func NewSeekerResponse(p models.SeekerProfile) SeekerResponse {
	response := SeekerResponse{
		ProfileResponse:       newProfileResponse(p.Profile),
		FirstName:             p.PersonalInfo.FirstName,
		LastName:              p.PersonalInfo.LastName,
		City:                  p.PersonalInfo.City,
		Phone:                 p.Contacts.Phone,
		Messenger:             p.Contacts.Messenger,
		ContactEmail:          p.Contacts.ContactEmail,
		PreferredPosition:     p.Preferences.PreferredPosition,
		PreferredCities:       p.Preferences.PreferredCities,
		WorkFormat:            p.Preferences.WorkFormat,
		CareerDetailsResponse: structured.NewCareerDetailsResponse(p.Career),
	}
	if p.PersonalInfo.AvatarKey != nil && *p.PersonalInfo.AvatarKey != "" {
		url := fmt.Sprintf("/api/seekers/%d/avatar", p.Profile.UserID)
		response.AvatarURL = &url
	}
	return response
}

type EmployerResponse struct {
	ProfileResponse
	CompanyName   string  `json:"company_name"`
	Industry      *string `json:"industry"`
	TeamSize      *string `json:"team_size"`
	Benefits      *string `json:"benefits"`
	Website       string  `json:"website"`
	HiringEmail   *string `json:"hiring_email"`
	MainCity      *string `json:"main_city"`
	OfficeAddress *string `json:"office_address"`
	structured.CompanyDetailsResponse
	LogoURL *string `json:"logo_url"`
}

func NewEmployerResponse(p models.EmployerProfile) EmployerResponse {
	response := EmployerResponse{
		ProfileResponse:        newProfileResponse(p.Profile),
		CompanyName:            p.Company.CompanyName,
		Industry:               p.Company.Industry,
		TeamSize:               p.Company.TeamSize,
		Benefits:               p.Company.Benefits,
		Website:                p.Contacts.Website,
		HiringEmail:            p.Contacts.HiringEmail,
		MainCity:               p.Location.MainCity,
		OfficeAddress:          p.Location.OfficeAddress,
		CompanyDetailsResponse: structured.NewCompanyDetailsResponse(p.Details),
	}
	if p.Company.LogoKey != nil && *p.Company.LogoKey != "" {
		url := fmt.Sprintf("/api/employers/%d/logo", p.Profile.UserID)
		response.LogoURL = &url
	}
	return response
}
