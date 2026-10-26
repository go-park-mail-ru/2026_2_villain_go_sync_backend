package models

import "time"

type Profile struct {
	UserID      int       `json:"user_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SeekerProfile struct {
	Profile
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type EmployerProfile struct {
	Profile
	CompanyName string `json:"company_name"`
	Website     string `json:"website"`
}
