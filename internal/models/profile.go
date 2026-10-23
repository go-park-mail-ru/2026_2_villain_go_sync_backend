package models

import "time"

type SeekerProfile struct {
	UserID    int       `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	About     string    `json:"about"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EmployerProfile struct {
	UserID      int       `json:"user_id"`
	CompanyName string    `json:"company_name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
