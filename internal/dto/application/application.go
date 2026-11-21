package application

import "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"

type CreateApplicationInput struct {
	SeekerID    int64
	VacancyID   int64
	ResumeID    int
	CoverLetter *string
}

type UpdateStatusInput struct {
	ApplicationID int64
	EmployerID    int64
	Status        models.ApplicationStatusValue
	ReasonCode    *models.RejectionReasonCode
	Comment       *string
}
