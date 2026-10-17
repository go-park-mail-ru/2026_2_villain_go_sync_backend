package vacancy

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type VacancyRepository interface {
	List(ctx context.Context) ([]models.Vacancy, error)
}

type VacancyUsecase struct {
	vacancies VacancyRepository
}

func NewVacancyUsecase(vacancies VacancyRepository) *VacancyUsecase {
	return &VacancyUsecase{vacancies: vacancies}
}

func (u *VacancyUsecase) List(ctx context.Context) ([]models.Vacancy, error) {
	return u.vacancies.List(ctx)
}
