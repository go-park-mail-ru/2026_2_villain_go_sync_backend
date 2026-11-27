package profile

import (
	seekerprofile "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/repository/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/repository/company"
)

type PgxPool = seekerprofile.PgxPool

type seekerRepository = seekerprofile.PostgresRepository

type PostgresRepository struct {
	*seekerRepository
	*company.PostgresRepository
}

func NewPostgresRepository(pool PgxPool) *PostgresRepository {
	return &PostgresRepository{
		seekerRepository:   seekerprofile.NewPostgresRepository(pool),
		PostgresRepository: company.NewPostgresRepository(pool),
	}
}
