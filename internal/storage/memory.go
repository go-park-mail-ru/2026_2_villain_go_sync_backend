package storage

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrEmailTaken = errors.New("email already taken")
	ErrNotFound   = errors.New("user not found")
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(user User) (User, error)
	GetByEmail(email string) (User, error)
}

type MemoryRepository struct {
	mu     sync.RWMutex
	users  map[string]User
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:  make(map[string]User),
		nextID: 1,
	}
}

func (m *MemoryRepository) Create(user User) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Email]; exists {
		return User{}, ErrEmailTaken
	}

	user.ID = m.nextID
	m.nextID++
	user.CreatedAt = time.Now()

	m.users[user.Email] = user

	return user, nil
}

func (m *MemoryRepository) GetByEmail(email string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[email]
	if !ok {
		return User{}, ErrNotFound
	}

	return user, nil
}
