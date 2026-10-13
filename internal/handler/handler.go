package handler

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type Handler struct {
	Storage   UserRepository
	Vacancies VacancyRepository
	Tokens    *auth.TokenManager
}

func NewHandler(
	storage UserRepository,
	vacancies VacancyRepository,
	tokens *auth.TokenManager,
) *Handler {
	return &Handler{
		Storage:   storage,
		Vacancies: vacancies,
		Tokens:    tokens,
	}
}

type UserRepository interface {
	Create(ctx context.Context, user models.User) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
}

type VacancyRepository interface {
	List(ctx context.Context) ([]models.Vacancy, error)
}
