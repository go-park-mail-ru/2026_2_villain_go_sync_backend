package applicant

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ApplicantRepository interface {
	IsOwner(ctx context.Context, vacancyID int, employerID int) (bool, error)
	ListByVacancyID(ctx context.Context, vacancyID int) ([]models.Applicant, error)
}

type ApplicantUsecase struct {
	applicants ApplicantRepository
}

func NewApplicantUsecase(applicants ApplicantRepository) *ApplicantUsecase {
	return &ApplicantUsecase{applicants: applicants}
}

func (u *ApplicantUsecase) List(ctx context.Context, userID int, role models.Role, vacancyID int) ([]models.Applicant, error) {
	if !role.IsEmployer() {
		return nil, apperrors.ErrForbidden
	}

	owned, err := u.applicants.IsOwner(ctx, vacancyID, userID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, apperrors.ErrForbidden
	}

	return u.applicants.ListByVacancyID(ctx, vacancyID)
}
