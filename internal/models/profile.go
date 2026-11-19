package models

import "time"

type Profile struct {
	UserID      int       `json:"user_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SeekerProfile struct {
	CareerDetails
	Profile
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	Phone             string  `json:"phone"`
	City              *string `json:"city"`
	Messenger         *string `json:"messenger"`
	ContactEmail      *string `json:"contact_email"`
	PreferredPosition *string `json:"preferred_position"`
	PreferredCities   *string `json:"preferred_cities"`
	WorkFormat        *string `json:"work_format"`
	AvatarKey         *string `json:"-"`
}

type EmployerProfile struct {
	CompanyDetails
	Profile
	CompanyName   string  `json:"company_name"`
	Website       string  `json:"website"`
	Industry      *string `json:"industry"`
	TeamSize      *string `json:"team_size"`
	MainCity      *string `json:"main_city"`
	Benefits      *string `json:"benefits"`
	HiringEmail   *string `json:"hiring_email"`
	OfficeAddress *string `json:"office_address"`
	LogoKey       *string `json:"-"`
}
