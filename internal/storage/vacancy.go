package storage

import (
	"sync"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type MemoryVacancyRepository struct {
	mu        sync.RWMutex
	vacancies []models.Vacancy
}

func NewMemoryVacancyRepository() *MemoryVacancyRepository {
	return &MemoryVacancyRepository{}
}

func (r *MemoryVacancyRepository) List() ([]models.Vacancy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.vacancies, nil
}
