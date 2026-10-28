package vacancy

import (
	"context"

	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type mockVacancyUsecase struct {
	listFn func(ctx context.Context) ([]models.Vacancy, error)
}

func (m *mockVacancyUsecase) ListVacancies(ctx context.Context) ([]models.Vacancy, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockVacancyUsecase) Create(ctx context.Context, input dto.CreateVacancyInput) (models.Vacancy, error) {
	return models.Vacancy{}, nil
}

func (m *mockVacancyUsecase) GetByID(ctx context.Context, id int64) (models.Vacancy, error) {
	return models.Vacancy{}, nil
}

func (m *mockVacancyUsecase) ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error) {
	return nil, nil
}

func (m *mockVacancyUsecase) Update(ctx context.Context, input dto.UpdateVacancyInput) (models.Vacancy, error) {
	return models.Vacancy{}, nil
}

func (m *mockVacancyUsecase) Delete(ctx context.Context, vacancyID, employerID int64) error {
	return nil
}

func (m *mockVacancyUsecase) ListApplicants(ctx context.Context, userID int64, role models.Role, vacancyID int64) ([]models.Applicant, error) {
	return nil, nil
}
