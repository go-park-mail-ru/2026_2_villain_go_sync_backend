package repository

import (
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/user"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	User    *user.PostgresRepository
	Vacancy *vacancy.MemoryRepository
}

func CreateConfig(pool *pgxpool.Pool) *Config {
	return &Config{
		User:    user.NewPostgresRepository(pool),
		Vacancy: vacancy.NewDemoRepository(),
	}
}
