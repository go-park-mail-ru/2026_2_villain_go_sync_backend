package vacancy

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func (u *VacancyUsecase) ListCategories(ctx context.Context) ([]models.Category, error) {
	return u.vacancies.ListCategories(ctx)
}
