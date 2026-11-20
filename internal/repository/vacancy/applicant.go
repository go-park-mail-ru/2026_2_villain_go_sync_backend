package vacancy

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
)

const (
	vacancyOwnerQuery = `SELECT EXISTS(SELECT 1 FROM vacancy WHERE vacancy_id = $1 AND employer_id = $2)`

	applicantsFrom = `
		  FROM application AS a
		  JOIN application_resume_snapshot AS snapshot ON snapshot.application_id = a.application_id
		  JOIN seeker_profile AS sp ON a.seeker_id = sp.user_id
		  JOIN vacancy_history AS vh ON vh.vacancy_id = a.vacancy_id AND vh.version = a.vacancy_version
		  JOIN vacancy AS v ON v.vacancy_id = a.vacancy_id
	`

	applicantsSelectQuery = `
		SELECT COALESCE(sp.first_name, ''), COALESCE(sp.last_name, ''),
		       COALESCE(latest_status.status, 'created'),
		       a.created_at, snapshot.title, a.application_id, a.vacancy_id, vh.title,
		       a.seeker_id, snapshot.source_resume_id, a.cover_letter,
		       COALESCE(sp.avatar_key, ''), a.updated_at, a.vacancy_version,
		       latest_status.status_history_id, latest_status.comment, latest_status.changed_at,
		       latest_status.changed_by, latest_status.reason_code
	` + applicantsFrom + `
		  LEFT JOIN (
		      SELECT DISTINCT ON (application_id)
		             application_id, status_history_id, status, comment, changed_at, changed_by, reason_code
		        FROM application_status_history
		       ORDER BY application_id, changed_at DESC, status_history_id DESC
		  ) AS latest_status ON latest_status.application_id = a.application_id
	`
	applicantsPage               = ` ORDER BY a.created_at DESC, a.application_id DESC LIMIT $2 OFFSET $3`
	vacancyApplicantsFilter      = ` WHERE a.vacancy_id = $1`
	employerApplicantsFilter     = ` WHERE v.employer_id = $1`
	vacancyApplicantsQuery       = applicantsSelectQuery + vacancyApplicantsFilter + applicantsPage
	employerApplicantsQuery      = applicantsSelectQuery + employerApplicantsFilter + applicantsPage
	countVacancyApplicantsQuery  = `SELECT COUNT(*)` + applicantsFrom + vacancyApplicantsFilter
	countEmployerApplicantsQuery = `SELECT COUNT(*)` + applicantsFrom + employerApplicantsFilter
)

func (r *PostgresRepository) IsOwner(ctx context.Context, vacancyID int64, employerID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, vacancyOwnerQuery, vacancyID, employerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check owner: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) ListByVacancyID(ctx context.Context, vacancyID int64, page, limit int) ([]models.Applicant, int64, error) {
	return r.listApplicants(ctx, vacancyApplicantsQuery, countVacancyApplicantsQuery, vacancyID, page, limit)
}

func (r *PostgresRepository) ListByEmployerID(ctx context.Context, employerID int64, page, limit int) ([]models.Applicant, int64, error) {
	return r.listApplicants(ctx, employerApplicantsQuery, countEmployerApplicantsQuery, employerID, page, limit)
}

func (r *PostgresRepository) listApplicants(ctx context.Context, query, countQuery string, id int64, page, limit int) ([]models.Applicant, int64, error) {
	var applicants []models.Applicant
	var total int64
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, countQuery, id).Scan(&total); err != nil {
			return fmt.Errorf("count applicants: %w", err)
		}
		var err error
		applicants, err = queryApplicants(ctx, tx, query, id, limit, int64(page-1)*int64(limit))
		return err
	})
	if err != nil {
		return nil, 0, err
	}
	return applicants, total, nil
}

func queryApplicants(ctx context.Context, tx pgx.Tx, query string, id int64, limit int, offset int64) ([]models.Applicant, error) {
	rows, err := tx.Query(ctx, query, id, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select applicants: %w", err)
	}
	defer rows.Close()

	applicants := make([]models.Applicant, 0)
	for rows.Next() {
		var app models.Applicant
		var avatarKey string
		err := rows.Scan(
			&app.FirstName,
			&app.LastName,
			&app.Status,
			&app.CreatedAt,
			&app.ResumeTitle,
			&app.ApplicationID,
			&app.VacancyID,
			&app.VacancyTitle,
			&app.SeekerID,
			&app.ResumeID,
			&app.CoverLetter,
			&avatarKey,
			&app.UpdatedAt,
			&app.VacancyVersion,
			&app.StatusHistoryID,
			&app.StatusComment,
			&app.StatusChangedAt,
			&app.StatusChangedBy,
			&app.ReasonCode,
		)
		if err != nil {
			return nil, fmt.Errorf("scan applicant: %w", err)
		}
		app.AppliedAt = app.CreatedAt
		app.SetAvatarURL(avatarKey)
		applicants = append(applicants, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return applicants, nil
}
