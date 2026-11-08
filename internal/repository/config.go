package repository

import (
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/s3"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/user"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	User        *user.PostgresRepository
	Vacancy     *vacancy.PostgresRepository
	Resume      *resume.PostgresRepository
	Profile     *profile.PostgresRepository
	Application *application.PostgresRepository
	S3          *s3.LoggingAdapter
}

func CreateConfig(pool *pgxpool.Pool, baseS3 *s3.Adapter) *Config {
	loggingPool := NewLoggingPool(pool)
	loggingS3 := s3.NewLoggingAdapter(baseS3)

	return &Config{
		User:        user.NewPostgresRepository(loggingPool),
		Vacancy:     vacancy.NewPostgresRepository(loggingPool),
		Resume:      resume.NewPostgresRepository(loggingPool),
		Profile:     profile.NewPostgresRepository(loggingPool),
		Application: application.NewPostgresRepository(loggingPool),
		S3:          loggingS3,
	}
}
