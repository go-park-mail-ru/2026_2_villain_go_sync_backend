package models

import "time"

type Role string

const (
	RoleSeeker   Role = "seeker"
	RoleEmployer Role = "employer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleSeeker, RoleEmployer:
		return true
	default:
		return false
	}
}

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
