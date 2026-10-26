package vacancy

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ApplicantRepository interface {
	IsOwner(ctx context.Context, vacancyID int64, employerID int64) (bool, error)
	ListByVacancyID(ctx context.Context, vacancyID int64) ([]models.Applicant, error)
}

type ApplicantUsecase struct {
	applicants ApplicantRepository
}

func NewApplicantUsecase(applicants ApplicantRepository) *ApplicantUsecase {
	return &ApplicantUsecase{applicants: applicants}
}

func (u *ApplicantUsecase) ListApplicants(ctx context.Context, userID int64, role models.Role, vacancyID int64) ([]models.Applicant, error) {
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
