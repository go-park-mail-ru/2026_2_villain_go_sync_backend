package vacancy

import (
	"context"
	"sync"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type MemoryRepository struct {
	mu        sync.RWMutex
	vacancies []models.Vacancy
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		vacancies: []models.Vacancy{},
	}
}

func (r *MemoryRepository) ListVacancies(ctx context.Context) ([]models.Vacancy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.vacancies, nil
}
