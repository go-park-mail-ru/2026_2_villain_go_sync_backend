package models

import "time"

type Profile struct {
	UserID      int
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SeekerProfile struct {
	Profile      Profile
	PersonalInfo SeekerPersonalInfo
	Contacts     SeekerContacts
	Preferences  SeekerPreferences
	Career       CareerDetails
}

type SeekerPersonalInfo struct {
	FirstName string
	LastName  string
	City      *string
	AvatarKey *string
}

type SeekerContacts struct {
	Phone        string
	Messenger    *string
	ContactEmail *string
}

type SeekerPreferences struct {
	PreferredPosition *string
	PreferredCities   *string
	WorkFormat        *WorkFormat
}

type EmployerProfile struct {
	Profile  Profile
	Company  EmployerCompanyInfo
	Contacts EmployerContacts
	Location EmployerLocation
	Details  CompanyDetails
}

type EmployerCompanyInfo struct {
	CompanyName string
	Industry    *string
	TeamSize    *string
	Benefits    *string
	LogoKey     *string
}

type EmployerContacts struct {
	Website     string
	HiringEmail *string
}

type EmployerLocation struct {
	MainCity      *string
	OfficeAddress *string
}
