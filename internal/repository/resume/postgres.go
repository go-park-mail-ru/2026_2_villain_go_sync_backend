package resume

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, resume models.Resume) (models.Resume, error) {
	query := `
		INSERT INTO resume (seeker_id, title, description, experience, education)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING resume_id, created_at, updated_at
	`
	err := r.pool.QueryRow(
		ctx, query, resume.SeekerID, resume.Title, resume.Description, resume.Experience, resume.Education,
	).Scan(&resume.ID, &resume.CreatedAt, &resume.UpdatedAt)

	if err != nil {
		return models.Resume{}, fmt.Errorf("insert resume: %w", err)
	}
	return resume, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int) (models.Resume, error) {
	query := `
		SELECT resume_id, seeker_id, title, description, experience, education, created_at, updated_at 
		FROM resume 
		WHERE resume_id = $1
	`
	var res models.Resume
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&res.ID, &res.SeekerID, &res.Title, &res.Description, &res.Experience, &res.Education, &res.CreatedAt, &res.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Resume{}, apperrors.ErrNotFound
	}
	if err != nil {
		return models.Resume{}, fmt.Errorf("select resume: %w", err)
	}
	return res, nil
}

func (r *PostgresRepository) ListBySeekerID(ctx context.Context, seekerID int) ([]models.Resume, error) {
	query := `
		SELECT resume_id, seeker_id, title, description, experience, education, created_at, updated_at 
		FROM resume 
		WHERE seeker_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, seekerID)
	if err != nil {
		return nil, fmt.Errorf("select resumes: %w", err)
	}
	defer rows.Close()

	var resumes []models.Resume
	for rows.Next() {
		var res models.Resume
		err := rows.Scan(&res.ID, &res.SeekerID, &res.Title, &res.Description, &res.Experience, &res.Education, &res.CreatedAt, &res.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan resume: %w", err)
		}
		resumes = append(resumes, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return resumes, nil
}

func (r *PostgresRepository) Update(ctx context.Context, resume models.Resume, seekerID int) (models.Resume, error) {
	query := `
		UPDATE resume 
		SET title = $1, description = $2, experience = $3, education = $4, updated_at = now()
		WHERE resume_id = $5 AND seeker_id = $6
		RETURNING updated_at
	`
	err := r.pool.QueryRow(
		ctx, query, resume.Title, resume.Description, resume.Experience, resume.Education, resume.ID, seekerID,
	).Scan(&resume.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Resume{}, apperrors.ErrNotFound
	}
	if err != nil {
		return models.Resume{}, fmt.Errorf("update resume: %w", err)
	}
	return resume, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int, seekerID int) error {
	query := `DELETE FROM resume WHERE resume_id = $1 AND seeker_id = $2`
	tag, err := r.pool.Exec(ctx, query, id, seekerID)
	if err != nil {
		return fmt.Errorf("delete resume: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.ErrNoContent
	}
	return nil
}
