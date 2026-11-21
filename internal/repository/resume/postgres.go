package resume

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/structured"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	foreignKeyViolationCode = "23503"
	resumeSeekerFk          = "resume_seeker_fk"

	resumeDetailsColumns = `target_position, city, employment_type, work_format, work_schedule, salary_from, portfolio_url, languages`

	resumeSelectQuery = `
		SELECT resume_id, seeker_id, title, COALESCE(description, ''),
		       COALESCE(l.experience_text, ''), COALESCE(l.education_text, ''),
		       r.created_at, r.updated_at,
		       ` + resumeDetailsColumns + `
		  FROM resume AS r
		  LEFT JOIN resume_legacy_content AS l USING (resume_id)
	`

	createQuery = `
		WITH created_resume AS (
		    INSERT INTO resume (
		        seeker_id, title, description, target_position, city, employment_type,
		        work_format, work_schedule, salary_from, portfolio_url, languages
		    )
		    VALUES ($1, $2, $3, $6, NULLIF($7, ''), NULLIF($8, ''),
		            NULLIF($9, ''), NULLIF($10, ''), $11, NULLIF($12, ''), NULLIF($13, ''))
		    RETURNING resume_id, created_at, updated_at,
		           ` + resumeDetailsColumns + `
		), saved_text AS (
		    INSERT INTO resume_legacy_content (resume_id, experience_text, education_text)
		    SELECT resume_id, $4, $5
		      FROM created_resume
		)
		SELECT resume_id, created_at, updated_at,
		           ` + resumeDetailsColumns + `
		  FROM created_resume
	`

	getByIDQuery = resumeSelectQuery + `
		 WHERE r.resume_id = $1 AND r.deleted_at IS NULL
	`

	listBySeekerIDQuery = resumeSelectQuery + `
		 WHERE r.seeker_id = $1 AND r.deleted_at IS NULL
		 ORDER BY r.created_at DESC
	`

	updateQuery = `
		WITH updated_resume AS (
		    UPDATE resume
		       SET title = $1,
		           description = $2,
		           target_position = $7,
		           city = NULLIF(COALESCE($8, city), ''),
		           employment_type = NULLIF(COALESCE($9, employment_type), ''),
		           work_format = NULLIF(COALESCE($10, work_format), ''),
		           work_schedule = NULLIF(COALESCE($11, work_schedule), ''),
		           salary_from = CASE WHEN $15 THEN NULL ELSE COALESCE($12, salary_from) END,
		           portfolio_url = NULLIF(COALESCE($13, portfolio_url), ''),
		           languages = NULLIF(COALESCE($14, languages), ''),
		           updated_at = NOW()
		     WHERE resume_id = $5 AND seeker_id = $6 AND deleted_at IS NULL
		    RETURNING resume_id, seeker_id, created_at, updated_at,
		           ` + resumeDetailsColumns + `
		), saved_text AS (
		    INSERT INTO resume_legacy_content (resume_id, experience_text, education_text)
		    SELECT resume_id, $3, $4
		      FROM updated_resume
		    ON CONFLICT (resume_id) DO UPDATE
		    SET experience_text = EXCLUDED.experience_text,
		        education_text = EXCLUDED.education_text
		)
		SELECT seeker_id, created_at, updated_at,
		           ` + resumeDetailsColumns + `
		  FROM updated_resume
	`

	deleteQuery = `DELETE FROM resume WHERE resume_id = $1 AND seeker_id = $2 AND deleted_at IS NULL`
)

type PgxPool interface {
	Begin(context.Context) (pgx.Tx, error)
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	pool PgxPool
}

func NewPostgresRepository(pool PgxPool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, resume models.Resume) (models.Resume, error) {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(
			ctx, createQuery, resume.SeekerID, resume.Title, resume.Content.Description, resume.Content.Experience, resume.Content.Education,
			resume.Preferences.TargetPosition, resume.Preferences.City, resume.Preferences.EmploymentType, resume.Preferences.WorkFormat,
			resume.Preferences.WorkSchedule, resume.Preferences.SalaryFrom, resume.Content.PortfolioURL, resume.Content.Languages,
		).Scan(
			&resume.ID, &resume.CreatedAt, &resume.UpdatedAt,
			&resume.Preferences.TargetPosition, &resume.Preferences.City, &resume.Preferences.EmploymentType, &resume.Preferences.WorkFormat,
			&resume.Preferences.WorkSchedule, &resume.Preferences.SalaryFrom, &resume.Content.PortfolioURL, &resume.Content.Languages,
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == foreignKeyViolationCode &&
					pgErr.ConstraintName == resumeSeekerFk {
					return apperrors.ErrNotFound
				}
			}
			return fmt.Errorf("insert resume: %w", err)
		}
		if err := structured.Resume.Replace(ctx, tx, resume.ID, resume.Career); err != nil {
			return err
		}
		return structured.Resume.Load(ctx, tx, map[int64]*models.CareerDetails{int64(resume.ID): &resume.Career})
	})
	if err != nil {
		return models.Resume{}, err
	}
	return resume, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int) (models.Resume, error) {
	var res models.Resume
	err := structured.ReadTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, getByIDQuery, id).
			Scan(
				&res.ID, &res.SeekerID, &res.Title, &res.Content.Description, &res.Content.Experience, &res.Content.Education, &res.CreatedAt, &res.UpdatedAt,
				&res.Preferences.TargetPosition, &res.Preferences.City, &res.Preferences.EmploymentType, &res.Preferences.WorkFormat,
				&res.Preferences.WorkSchedule, &res.Preferences.SalaryFrom, &res.Content.PortfolioURL, &res.Content.Languages,
			)

		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("select resume: %w", err)
		}
		return structured.Resume.Load(ctx, tx, map[int64]*models.CareerDetails{int64(res.ID): &res.Career})
	})
	if err != nil {
		return models.Resume{}, err
	}
	return res, nil
}

func (r *PostgresRepository) ListBySeekerID(ctx context.Context, seekerID int) ([]models.Resume, error) {
	resumes := make([]models.Resume, 0)
	err := structured.ReadTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, listBySeekerIDQuery, seekerID)
		if err != nil {
			return fmt.Errorf("select resumes: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var res models.Resume
			err := rows.Scan(
				&res.ID, &res.SeekerID, &res.Title, &res.Content.Description, &res.Content.Experience, &res.Content.Education, &res.CreatedAt, &res.UpdatedAt,
				&res.Preferences.TargetPosition, &res.Preferences.City, &res.Preferences.EmploymentType, &res.Preferences.WorkFormat,
				&res.Preferences.WorkSchedule, &res.Preferences.SalaryFrom, &res.Content.PortfolioURL, &res.Content.Languages,
			)
			if err != nil {
				return fmt.Errorf("scan resume: %w", err)
			}
			resumes = append(resumes, res)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("rows iteration error: %w", err)
		}

		rows.Close()
		targets := make(map[int64]*models.CareerDetails, len(resumes))
		for i := range resumes {
			targets[int64(resumes[i].ID)] = &resumes[i].Career
		}
		return structured.Resume.Load(ctx, tx, targets)
	})
	if err != nil {
		return nil, err
	}
	return resumes, nil
}

const getOwnerIDQuery = `
	SELECT seeker_id
	  FROM resume
	 WHERE resume_id = $1 AND deleted_at IS NULL
`

func (r *PostgresRepository) GetOwnerID(ctx context.Context, id int) (int, error) {
	var ownerID int
	err := r.pool.QueryRow(ctx, getOwnerIDQuery, id).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, apperrors.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get resume owner: %w", err)
	}
	return ownerID, nil
}

func (r *PostgresRepository) Update(ctx context.Context, input dto.UpdateInput, seekerID int) (models.Resume, error) {
	resume := input.Resume
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(
			ctx, updateQuery, resume.Title, resume.Content.Description, resume.Content.Experience, resume.Content.Education, resume.ID, seekerID,
			resume.Preferences.TargetPosition, resume.Preferences.City, resume.Preferences.EmploymentType, resume.Preferences.WorkFormat,
			resume.Preferences.WorkSchedule, resume.Preferences.SalaryFrom, resume.Content.PortfolioURL, resume.Content.Languages, input.ClearSalaryFrom,
		).Scan(
			&resume.SeekerID, &resume.CreatedAt, &resume.UpdatedAt,
			&resume.Preferences.TargetPosition, &resume.Preferences.City, &resume.Preferences.EmploymentType, &resume.Preferences.WorkFormat,
			&resume.Preferences.WorkSchedule, &resume.Preferences.SalaryFrom, &resume.Content.PortfolioURL, &resume.Content.Languages,
		)

		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("update resume: %w", err)
		}
		if err := structured.Resume.Replace(ctx, tx, resume.ID, resume.Career); err != nil {
			return err
		}
		return structured.Resume.Load(ctx, tx, map[int64]*models.CareerDetails{int64(resume.ID): &resume.Career})
	})
	if err != nil {
		return models.Resume{}, err
	}
	return resume, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int, seekerID int) error {
	tag, err := r.pool.Exec(ctx, deleteQuery, id, seekerID)
	if err != nil {
		return fmt.Errorf("delete resume: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.ErrNoContent
	}
	return nil
}
