package profile

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

func (r *PostgresRepository) GetSeeker(ctx context.Context, userID int) (models.SeekerProfile, error) {
	query := `
		SELECT user_id, first_name, last_name, phone, about, created_at, updated_at 
		FROM seeker_profile 
		WHERE user_id = $1
	`
	var p models.SeekerProfile
	err := r.pool.QueryRow(ctx, query, userID).Scan(&p.UserID, &p.FirstName, &p.LastName, &p.Phone, &p.About, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.SeekerProfile{}, apperrors.ErrNotFound
	}
	return p, err
}

func (r *PostgresRepository) UpsertSeeker(ctx context.Context, p models.SeekerProfile) (models.SeekerProfile, error) {
	query := `
		INSERT INTO seeker_profile (user_id, first_name, last_name, phone, about)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE 
		SET first_name = EXCLUDED.first_name, 
			last_name = EXCLUDED.last_name, 
			phone = EXCLUDED.phone, 
			about = EXCLUDED.about, 
			updated_at = now()
		RETURNING created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, p.UserID, p.FirstName, p.LastName, p.Phone, p.About).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.SeekerProfile{}, fmt.Errorf("upsert seeker: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) GetEmployer(ctx context.Context, userID int) (models.EmployerProfile, error) {
	query := `
		SELECT user_id, company_name, description, website, created_at, updated_at 
		FROM employer_profile
		WHERE user_id = $1
	`
	var p models.EmployerProfile
	err := r.pool.QueryRow(ctx, query, userID).Scan(&p.UserID, &p.CompanyName, &p.Description, &p.Website, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.EmployerProfile{}, apperrors.ErrNotFound
	}
	return p, err
}

func (r *PostgresRepository) UpsertEmployer(ctx context.Context, p models.EmployerProfile) (models.EmployerProfile, error) {
	query := `
		INSERT INTO employer_profile (user_id, company_name, description, website)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE 
		SET company_name = EXCLUDED.company_name, 
			description = EXCLUDED.description, 
			website = EXCLUDED.website, 
			updated_at = now()
		RETURNING created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, p.UserID, p.CompanyName, p.Description, p.Website).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.EmployerProfile{}, fmt.Errorf("upsert employer: %w", err)
	}
	return p, nil
}
