package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertApplicationQuery = `
		INSERT INTO application (
			vacancy_id, vacancy_version, resume_id, cover_letter
		)
		VALUES ($1, $2, $3, $4)
		RETURNING application_id, created_at, updated_at
	`

	insertInitialStatusQuery = `
		INSERT INTO application_status_history (application_id, status)
		VALUES ($1, 'created')
	`

	uniqueViolationCode         = "23505"
	foreignKeyViolationCode     = "23503"
	applicationVacancyResumeKey = "application_vacancy_resume_unique"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	application models.Application,
) (models.Application, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Application{}, fmt.Errorf("begin application transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		insertApplicationQuery,
		application.VacancyID,
		application.VacancyVersion,
		application.ResumeID,
		application.CoverLetter,
	).Scan(
		&application.ID,
		&application.CreatedAt,
		&application.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueViolationCode &&
				pgErr.ConstraintName == applicationVacancyResumeKey {
				return models.Application{}, apperrors.ErrConflict
			}

			if pgErr.Code == foreignKeyViolationCode {
				return models.Application{}, apperrors.ErrNotFound
			}
		}

		return models.Application{}, fmt.Errorf("insert application: %w", err)
	}

	_, err = tx.Exec(ctx, insertInitialStatusQuery, application.ID)
	if err != nil {
		return models.Application{}, fmt.Errorf("insert initial application status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Application{}, fmt.Errorf("commit application transaction: %w", err)
	}

	return application, nil
}
