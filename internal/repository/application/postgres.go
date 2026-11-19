package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	insertApplicationQuery = `
		SELECT new_application_id, created_on, updated_on
		FROM create_application($1, $2, $3, $4, $5)
	`

	uniqueViolationCode         = "23505"
	foreignKeyViolationCode     = "23503"
	applicationVacancySeekerKey = "application_vacancy_seeker_unique"
)

type PgxPool interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresRepository struct {
	pool PgxPool
}

func NewPostgresRepository(pool PgxPool) *PostgresRepository {
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
		application.SeekerID,
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
				pgErr.ConstraintName == applicationVacancySeekerKey {
				return models.Application{}, apperrors.ErrConflict
			}

			if pgErr.Code == "23514" {
				return models.Application{}, apperrors.ErrConflict
			}

			if pgErr.Code == foreignKeyViolationCode {
				return models.Application{}, apperrors.ErrNotFound
			}
		}

		return models.Application{}, fmt.Errorf("insert application: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Application{}, fmt.Errorf("commit application transaction: %w", err)
	}

	return application, nil
}

const updateStatusQuery = `
	WITH updated_application AS (
	    UPDATE application AS a
	       SET updated_at = GREATEST(a.updated_at, clock_timestamp())
	      FROM vacancy AS v
	     WHERE a.application_id = $1
	       AND v.vacancy_id = a.vacancy_id
	       AND v.employer_id = $2
	    RETURNING a.application_id, a.updated_at
	)
	INSERT INTO application_status_history (
	    application_id, status, reason_code, comment, changed_by, changed_at
	)
	SELECT application_id, $3, $4, $5, $2, updated_at
	  FROM updated_application
	RETURNING application_id, status, reason_code, comment, changed_at
`

func (r *PostgresRepository) UpdateStatus(ctx context.Context, input dto.UpdateStatusInput) (models.ApplicationStatusUpdate, error) {
	var result models.ApplicationStatusUpdate
	err := r.pool.QueryRow(ctx, updateStatusQuery,
		input.ApplicationID, input.EmployerID, input.Status, input.ReasonCode, input.Comment,
	).Scan(&result.ApplicationID, &result.Status, &result.ReasonCode, &result.StatusComment, &result.StatusChangedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ApplicationStatusUpdate{}, apperrors.ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return models.ApplicationStatusUpdate{}, apperrors.ErrInvalidInput
		}
		return models.ApplicationStatusUpdate{}, fmt.Errorf("record application status: %w", err)
	}
	return result, nil
}

const listBySeekerQuery = `
	SELECT a.application_id, a.seeker_id, a.vacancy_id, a.vacancy_version,
	       a.cover_letter, a.created_at, a.updated_at, snapshot.source_resume_id,
	       vh.title, ep.company_name, snapshot.title,
	       COALESCE(latest.status, 'created'),
	       latest.reason_code, latest.comment, COALESCE(latest.changed_at, a.created_at)
	  FROM application AS a
	  JOIN application_resume_snapshot AS snapshot ON snapshot.application_id = a.application_id
	  JOIN vacancy_history AS vh ON vh.vacancy_id = a.vacancy_id AND vh.version = a.vacancy_version
	  JOIN vacancy AS v ON v.vacancy_id = a.vacancy_id
	  JOIN employer_profile AS ep ON ep.user_id = v.employer_id
	  LEFT JOIN LATERAL (
	      SELECT status, reason_code, comment, changed_at
	        FROM application_status_history
	       WHERE application_id = a.application_id
	       ORDER BY changed_at DESC, status_history_id DESC
	       LIMIT 1
	  ) AS latest ON TRUE
	 WHERE a.seeker_id = $1
	 ORDER BY a.created_at DESC, a.application_id DESC
`

func (r *PostgresRepository) ListBySeeker(ctx context.Context, seekerID int64) ([]models.SeekerApplication, error) {
	rows, err := r.pool.Query(ctx, listBySeekerQuery, seekerID)
	if err != nil {
		return nil, fmt.Errorf("list seeker applications: %w", err)
	}
	defer rows.Close()

	result := make([]models.SeekerApplication, 0)
	for rows.Next() {
		var app models.SeekerApplication
		if err := rows.Scan(
			&app.ID, &app.SeekerID, &app.VacancyID, &app.VacancyVersion,
			&app.CoverLetter, &app.CreatedAt, &app.UpdatedAt, &app.ResumeID,
			&app.VacancyTitle, &app.CompanyName, &app.ResumeTitle,
			&app.Status, &app.ReasonCode, &app.StatusComment, &app.StatusChangedAt,
		); err != nil {
			return nil, fmt.Errorf("scan seeker application: %w", err)
		}
		result = append(result, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list seeker applications: %w", err)
	}
	return result, nil
}
