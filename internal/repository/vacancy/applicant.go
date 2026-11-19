package vacancy

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

const (
	vacancyOwnerQuery = `SELECT EXISTS(SELECT 1 FROM vacancy WHERE vacancy_id = $1 AND employer_id = $2)`

	applicantsSelectQuery = `
		SELECT COALESCE(sp.first_name, ''), COALESCE(sp.last_name, ''),
		       COALESCE(latest_status.status, 'created'),
		       a.created_at, snapshot.title, a.application_id, a.vacancy_id, vh.title,
		       a.seeker_id, snapshot.source_resume_id, a.cover_letter,
		       COALESCE(sp.avatar_key, ''), a.updated_at, a.vacancy_version,
		       latest_status.status_history_id, latest_status.comment, latest_status.changed_at,
		       latest_status.changed_by, latest_status.reason_code
		  FROM application AS a
		  JOIN application_resume_snapshot AS snapshot ON snapshot.application_id = a.application_id
		  JOIN seeker_profile AS sp ON a.seeker_id = sp.user_id
		  JOIN vacancy_history AS vh ON vh.vacancy_id = a.vacancy_id AND vh.version = a.vacancy_version
		  JOIN vacancy AS v ON v.vacancy_id = a.vacancy_id
		  LEFT JOIN (
		      SELECT DISTINCT ON (application_id)
		             application_id, status_history_id, status, comment, changed_at, changed_by, reason_code
		        FROM application_status_history
		       ORDER BY application_id, changed_at DESC, status_history_id DESC
		  ) AS latest_status ON latest_status.application_id = a.application_id
	`

	applicantsOrder         = ` ORDER BY a.created_at DESC, a.application_id DESC`
	vacancyApplicantsQuery  = applicantsSelectQuery + ` WHERE a.vacancy_id = $1` + applicantsOrder
	employerApplicantsQuery = applicantsSelectQuery + ` WHERE v.employer_id = $1` + applicantsOrder
)

func (r *PostgresRepository) IsOwner(ctx context.Context, vacancyID int64, employerID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, vacancyOwnerQuery, vacancyID, employerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check owner: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) ListByVacancyID(ctx context.Context, vacancyID int64) ([]models.Applicant, error) {
	return r.listApplicants(ctx, vacancyApplicantsQuery, vacancyID)
}

func (r *PostgresRepository) ListByEmployerID(ctx context.Context, employerID int64) ([]models.Applicant, error) {
	return r.listApplicants(ctx, employerApplicantsQuery, employerID)
}

func (r *PostgresRepository) listApplicants(ctx context.Context, query string, id int64) ([]models.Applicant, error) {
	rows, err := r.pool.Query(ctx, query, id)
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
