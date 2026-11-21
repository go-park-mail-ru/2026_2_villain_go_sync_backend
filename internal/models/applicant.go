package models

import (
	"fmt"
	"time"
)

type Applicant struct {
	ApplicationID int64
	CoverLetter   *string
	AppliedAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Seeker  ApplicantSeeker
	Vacancy ApplicantVacancy
	Resume  ApplicantResume
	Status  ApplicantStatus
}

type ApplicantSeeker struct {
	SeekerID  int64
	FirstName string
	LastName  string
	AvatarURL *string
}

type ApplicantVacancy struct {
	VacancyID      int64
	VacancyTitle   string
	VacancyVersion int
}

type ApplicantResume struct {
	ResumeID    *int64
	ResumeTitle string
}

type ApplicantStatus struct {
	Value           ApplicationStatusValue
	StatusHistoryID *int64
	StatusComment   *string
	StatusChangedAt *time.Time
	StatusChangedBy *int64
	ReasonCode      *RejectionReasonCode
}

func (a *Applicant) SetAvatarURL(avatarKey string) {
	if avatarKey == "" {
		a.Seeker.AvatarURL = nil
		return
	}

	url := fmt.Sprintf("/api/seekers/%d/avatar", a.Seeker.SeekerID)
	a.Seeker.AvatarURL = &url
}
