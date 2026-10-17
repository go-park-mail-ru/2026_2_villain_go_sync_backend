package vacancy

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type mockVacancyUsecase struct {
	listFn func(ctx context.Context) ([]models.Vacancy, error)
}

func (m *mockVacancyUsecase) List(ctx context.Context) ([]models.Vacancy, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}
