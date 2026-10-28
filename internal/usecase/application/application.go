package application

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ApplicationRepository interface {
	Create(ctx context.Context, input models.Application) (models.Application, error)
}

type ResumeRepository interface {
	GetByID(ctx context.Context, id int) (models.Resume, error)
}

type VacancyRepository interface {
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
}

type UserRepository interface {
	GetRoleByID(ctx context.Context, id int64) (string, error)
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
	if input.SeekerID <= 0 || input.VacancyID <= 0 || input.ResumeID <= 0 {
		return models.Application{}, apperrors.ErrInvalidInput
	}

	role, err := u.users.GetRoleByID(ctx, input.SeekerID)
	if err != nil {
		return models.Application{}, fmt.Errorf("get user role: %w", err)
	}
	if role != "seeker" {
		return models.Application{}, apperrors.ErrForbidden
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
		VacancyID:      vacancy.ID,
		VacancyVersion: vacancy.Version,
		ResumeID:       int64(input.ResumeID),
		CoverLetter:    input.CoverLetter,
	})
}
