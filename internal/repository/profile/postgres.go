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

func (r *PostgresRepository) GetSeekerByID(ctx context.Context, userID int) (models.SeekerProfile, error) {
	query := `
		SELECT user_id, first_name, last_name, phone, about, avatar_key, created_at, updated_at
		FROM seeker_profile 
		WHERE user_id = $1
	`
	var p models.SeekerProfile
	err := r.pool.QueryRow(ctx, query, userID).Scan(&p.UserID, &p.FirstName, &p.LastName, &p.Phone, &p.Description, &p.AvatarKey, &p.CreatedAt, &p.UpdatedAt)
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
		RETURNING created_at, updated_at, avatar_key

	`
	err := r.pool.QueryRow(ctx, query, p.UserID, p.FirstName, p.LastName, p.Phone, p.Description).Scan(&p.CreatedAt, &p.UpdatedAt, &p.AvatarKey)
	if err != nil {
		return models.SeekerProfile{}, fmt.Errorf("upsert seeker: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) GetEmployerByID(ctx context.Context, userID int) (models.EmployerProfile, error) {
	query := `
		SELECT user_id, company_name, description, website, logo_key, created_at, updated_at
		FROM employer_profile
		WHERE user_id = $1
	`
	var p models.EmployerProfile
	err := r.pool.QueryRow(ctx, query, userID).Scan(&p.UserID, &p.CompanyName, &p.Description, &p.Website, &p.LogoKey, &p.CreatedAt, &p.UpdatedAt)
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
		RETURNING created_at, updated_at, logo_key
	`
	err := r.pool.QueryRow(ctx, query, p.UserID, p.CompanyName, p.Description, p.Website).Scan(&p.CreatedAt, &p.UpdatedAt, &p.LogoKey)
	if err != nil {
		return models.EmployerProfile{}, fmt.Errorf("upsert employer: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) UpdateSeekerAvatarKey(ctx context.Context, userID int, key *string) error {
	query := `UPDATE seeker_profile
	SET avatar_key=$1
	WHERE user_id=$2`
	tag, err := r.pool.Exec(ctx, query, key, userID)
	if err != nil {
		return fmt.Errorf("update avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound

	}
	return nil
}

func (r *PostgresRepository) UpdateEmployerLogoKey(ctx context.Context, userID int, key *string) error {
	query := `UPDATE employer_profile
	SET logo_key=$1
	WHERE user_id=$2`
	tag, err := r.pool.Exec(ctx, query, key, userID)
	if err != nil {
		return fmt.Errorf("update logo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound

	}
	return nil
}
