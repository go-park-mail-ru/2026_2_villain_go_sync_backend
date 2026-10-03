package handler

import "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"

type mockVacancyRepository struct {
	vacancies []models.Vacancy
	err       error
	calls     int
}

func (m *mockVacancyRepository) List() ([]models.Vacancy, error) {
	m.calls++
	return m.vacancies, m.err
}
