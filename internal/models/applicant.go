package models

import "time"

type Applicant struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Status      string    `json:"status"`
	AppliedAt   time.Time `json:"applied_at"`
	ResumeTitle string    `json:"resume_title"`
}
