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

func (r Role) IsSeeker() bool {
	return r == RoleSeeker
}

func (r Role) IsEmployer() bool {
	return r == RoleEmployer
}

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Applicant struct {
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Status        string    `json:"status"`
	AppliedAt     time.Time `json:"applied_at"`
	ResumeTitle   string    `json:"resume_title"`
	ApplicationID int64     `json:"application_id"`
	VacancyID     int64     `json:"vacancy_id"`
	VacancyTitle  string    `json:"vacancy_title"`
	SeekerID      int64     `json:"seeker_id"`
	ResumeID      int64     `json:"resume_id"`
	CoverLetter   *string   `json:"cover_letter"`
	AvatarURL     *string   `json:"avatar_url"`
}
