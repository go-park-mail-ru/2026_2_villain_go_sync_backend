package vacancy

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func (u *VacancyUsecase) ListApplicants(ctx context.Context, userID int64, role models.Role, vacancyID int64, page int, limit int) ([]models.Applicant, int64, error) {
	if !role.IsEmployer() {
		return nil, 0, apperrors.ErrForbidden
	}

	owned, err := u.vacancies.IsOwner(ctx, vacancyID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !owned {
		return nil, 0, apperrors.ErrForbidden
	}

	return u.vacancies.ListByVacancyID(ctx, vacancyID, page, limit)
}
