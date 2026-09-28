package storage

import (
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	users  map[string]models.User
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:  make(map[string]models.User),
		nextID: 1,
	}
}

func (m *MemoryRepository) Create(user models.User) (models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Email]; exists {
		return models.User{}, apperrors.ErrEmailTaken
	}

	user.ID = m.nextID
	m.nextID++
	user.CreatedAt = time.Now()

	m.users[user.Email] = user

	return user, nil
}

func (m *MemoryRepository) GetByEmail(email string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[email]
	if !ok {
		return models.User{}, apperrors.ErrNotFound
	}

	return user, nil
}
