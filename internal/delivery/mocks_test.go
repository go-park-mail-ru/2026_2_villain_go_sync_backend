package delivery

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type mockVacancyRepository struct {
	vacancies []models.Vacancy
	err       error
	calls     int
}

func (m *mockVacancyRepository) List(ctx context.Context) ([]models.Vacancy, error) {
	m.calls++
	return m.vacancies, m.err
}

type mockUserRepository struct {
	user            models.User
	createErr       error
	getByEmailErr   error
	createdUser     models.User
	requestedEmail  string
	createCalls     int
	getByEmailCalls int
}

func (m *mockUserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	m.createCalls++
	m.createdUser = user
	return m.user, m.createErr
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	m.getByEmailCalls++
	m.requestedEmail = email
	return m.user, m.getByEmailErr
}
