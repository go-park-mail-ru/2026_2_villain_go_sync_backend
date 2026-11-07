package app

import (
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/s3"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Repositories *repository.Config
	Usecases     *usecase.Config
	Handlers     *delivery.Config
}

func CreateConfig(pool *pgxpool.Pool, tokens *auth.TokenManager, s3Adapter *s3.Adapter) *Config {
	repos := repository.CreateConfig(pool, s3Adapter)
	usecases := usecase.CreateConfig(repos, tokens)
	handlers := delivery.CreateConfig(usecases)

	return &Config{
		Repositories: repos,
		Usecases:     usecases,
		Handlers:     handlers,
	}
}
