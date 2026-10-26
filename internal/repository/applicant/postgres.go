package applicant

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) IsOwner(ctx context.Context, vacancyID int, employerID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM vacancy WHERE vacancy_id = $1 AND employer_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, vacancyID, employerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check owner: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) ListByVacancyID(ctx context.Context, vacancyID int) ([]models.Applicant, error) {
	query := `
		SELECT 
			sp.first_name, 
			sp.last_name, 
			COALESCE(latest_status.status, 'pending') AS status, 
			a.created_at AS applied_at, 
			r.title AS resume_title
		FROM application a
		JOIN resume r ON a.resume_id = r.resume_id
		JOIN seeker_profile sp ON r.seeker_id = sp.user_id
		LEFT JOIN (
			SELECT DISTINCT ON (application_id) application_id, status
			FROM application_status_history
			ORDER BY application_id, changed_at DESC
		) latest_status ON a.application_id = latest_status.application_id
		WHERE a.vacancy_id = $1
		ORDER BY a.created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("select applicants: %w", err)
	}
	defer rows.Close()

	var applicants []models.Applicant
	for rows.Next() {
		var app models.Applicant
		err := rows.Scan(
			&app.FirstName,
			&app.LastName,
			&app.Status,
			&app.AppliedAt,
			&app.ResumeTitle,
		)
		if err != nil {
			return nil, fmt.Errorf("scan applicant: %w", err)
		}
		applicants = append(applicants, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return applicants, nil
}
