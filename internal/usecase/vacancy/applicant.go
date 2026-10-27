package vacancy

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func (u *VacancyUsecase) ListApplicants(ctx context.Context, userID int64, role models.Role, vacancyID int64) ([]models.Applicant, error) {
	if !role.IsEmployer() {
		return nil, apperrors.ErrForbidden
	}

	owned, err := u.vacancies.IsOwner(ctx, vacancyID, userID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, apperrors.ErrForbidden
	}

	return u.vacancies.ListByVacancyID(ctx, vacancyID)
}
