package application

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

//go:generate mockgen -source=application.go -destination=mock_repository.go -package=application
type ApplicationRepository interface {
	GetVacancyEmployerID(ctx context.Context, applicationID int64) (int64, error)
	UpdateStatus(ctx context.Context, input dto.UpdateStatusInput) (models.ApplicationStatusUpdate, error)
	ListBySeeker(ctx context.Context, seekerID int64, page int, limit int) ([]models.SeekerApplication, int64, error)
	Create(ctx context.Context, input models.Application) (models.Application, error)
}

type ResumeRepository interface {
	GetByID(ctx context.Context, id int) (models.Resume, error)
}

type VacancyRepository interface {
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
}

type UserRepository interface {
	GetRoleByID(ctx context.Context, id int64) (models.Role, error)
}

type ApplicationUsecase struct {
	applications ApplicationRepository
	resumes      ResumeRepository
	vacancies    VacancyRepository
	users        UserRepository
}

func NewApplicationUsecase(applications ApplicationRepository, resumes ResumeRepository, vacancies VacancyRepository, users UserRepository) *ApplicationUsecase {
	return &ApplicationUsecase{
		applications: applications,
		resumes:      resumes,
		vacancies:    vacancies,
		users:        users,
	}
}

func (u *ApplicationUsecase) Create(
	ctx context.Context,
	input dto.CreateApplicationInput,
) (models.Application, error) {
	if input.SeekerID <= 0 {
		return models.Application{}, apperrors.ErrInvalidInput
	}

	role, err := u.users.GetRoleByID(ctx, input.SeekerID)
	if err != nil {
		return models.Application{}, fmt.Errorf("get user role: %w", err)
	}
	if !role.IsSeeker() {
		return models.Application{}, apperrors.ErrForbidden
	}

	if input.VacancyID <= 0 || input.ResumeID <= 0 {
		return models.Application{}, apperrors.ErrInvalidInput
	}

	resume, err := u.resumes.GetByID(ctx, input.ResumeID)
	if err != nil {
		return models.Application{}, fmt.Errorf("get resume: %w", err)
	}

	if int64(resume.SeekerID) != input.SeekerID {
		return models.Application{}, apperrors.ErrForbidden
	}

	vacancy, err := u.vacancies.GetByID(ctx, input.VacancyID)
	if err != nil {
		return models.Application{}, fmt.Errorf("get vacancy: %w", err)
	}

	return u.applications.Create(ctx, models.Application{
		SeekerID:       input.SeekerID,
		VacancyID:      vacancy.ID,
		VacancyVersion: vacancy.Version,
		ResumeID:       int64(input.ResumeID),
		CoverLetter:    input.CoverLetter,
	})
}

func (u *ApplicationUsecase) UpdateStatus(ctx context.Context, role models.Role, input dto.UpdateStatusInput) (models.ApplicationStatusUpdate, error) {
	if !role.IsEmployer() {
		return models.ApplicationStatusUpdate{}, apperrors.ErrForbidden
	}
	if input.EmployerID <= 0 || input.ApplicationID <= 0 {
		return models.ApplicationStatusUpdate{}, apperrors.ErrInvalidInput
	}

	ownerID, err := u.applications.GetVacancyEmployerID(ctx, input.ApplicationID)
	if err != nil {
		return models.ApplicationStatusUpdate{}, err
	}
	if ownerID != input.EmployerID {
		return models.ApplicationStatusUpdate{}, apperrors.ErrNotOwner
	}

	if !input.Status.CanSetByEmployer() {
		return models.ApplicationStatusUpdate{}, fmt.Errorf("%w: invalid status", apperrors.ErrInvalidInput)
	}
	if input.Status == models.ApplicationRejected {
		if input.ReasonCode == nil {
			return models.ApplicationStatusUpdate{}, fmt.Errorf("%w: reason_code is required for rejected", apperrors.ErrInvalidInput)
		}
		if !input.ReasonCode.IsValid() {
			return models.ApplicationStatusUpdate{}, fmt.Errorf("%w: invalid reason_code", apperrors.ErrInvalidInput)
		}
	} else if input.ReasonCode != nil {
		return models.ApplicationStatusUpdate{}, fmt.Errorf("%w: reason_code is only allowed for rejected", apperrors.ErrInvalidInput)
	}

	return u.applications.UpdateStatus(ctx, input)
}

func (u *ApplicationUsecase) ListBySeeker(ctx context.Context, seekerID int64, role models.Role, page int, limit int) ([]models.SeekerApplication, int64, error) {
	if !role.IsSeeker() {
		return nil, 0, apperrors.ErrForbidden
	}
	return u.applications.ListBySeeker(ctx, seekerID, page, limit)
}
