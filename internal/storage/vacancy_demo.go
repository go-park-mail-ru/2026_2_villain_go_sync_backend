package storage

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func NewDemoVacancyRepository() *MemoryVacancyRepository {
	repo := NewMemoryVacancyRepository()
	createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)

	repo.vacancies = []models.Vacancy{
		{
			ID:          1,
			EmployerID:  1,
			Title:       "Go-разработчик",
			Description: "Разработка backend-сервисов на Go. Работа с HTTP API и PostgreSQL.",
			SalaryFrom:  120000,
			SalaryTo:    180000,
			CreatedAt:   createdAt,
		},
		{
			ID:          2,
			EmployerID:  2,
			Title:       "Frontend-разработчик",
			Description: "Разработка интерфейсов на JavaScript. Интеграция с API.",
			SalaryFrom:  100000,
			SalaryTo:    160000,
			CreatedAt:   createdAt,
		},
		{
			ID:          3,
			EmployerID:  1,
			Title:       "Стажёр DevOps",
			Description: "Автоматизация сборки и развёртывания. Работа с Linux и Docker.",
			SalaryFrom:  50000,
			SalaryTo:    80000,
			CreatedAt:   createdAt,
		},
	}

	return repo
}
