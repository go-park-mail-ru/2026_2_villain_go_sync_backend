package delivery

import (
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase"
)

type Config struct {
	Auth    *auth.Handler
	Vacancy *vacancy.Handler
}

func CreateConfig(usecases *usecase.Config) *Config {
	return &Config{
		Auth:    auth.NewHandler(usecases.Auth),
		Vacancy: vacancy.NewHandler(usecases.Vacancy),
	}
}
