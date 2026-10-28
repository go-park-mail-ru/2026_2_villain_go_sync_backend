package vacancy

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func float64Ptr(v float64) *float64 {
	return &v
}

func NewDemoRepository() *MemoryRepository {
	repo := NewMemoryRepository()
	createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)

	repo.vacancies = []models.Vacancy{
		{
			ID:          1,
			EmployerID:  1,
			Title:       "Go-разработчик",
			Description: "Разработка backend-сервисов на Go. Работа с HTTP API и PostgreSQL.",
			SalaryFrom:  float64Ptr(120000),
			SalaryTo:    float64Ptr(180000),
			CreatedAt:   createdAt,
		},
		{
			ID:          2,
			EmployerID:  2,
			Title:       "Frontend-разработчик",
			Description: "Разработка интерфейсов на JavaScript. Интеграция с API.",
			SalaryFrom:  float64Ptr(100000),
			SalaryTo:    float64Ptr(160000),
			CreatedAt:   createdAt,
		},
		{
			ID:          3,
			EmployerID:  1,
			Title:       "Стажёр DevOps",
			Description: "Автоматизация сборки и развёртывания. Работа с Linux и Docker.",
			SalaryFrom:  float64Ptr(50000),
			SalaryTo:    float64Ptr(80000),
			CreatedAt:   createdAt,
		},
	}

	return repo
}
