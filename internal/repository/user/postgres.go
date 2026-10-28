package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	query := `
        INSERT INTO app_user (email, password_hash, role)
        VALUES ($1, $2, $3)
        RETURNING user_id, created_at
    `

	err := r.pool.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, apperrors.ErrEmailTaken
		}
		return models.User{}, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	query := `
        SELECT user_id, email, password_hash, role, created_at
        FROM app_user
        WHERE email = $1
    `

	var user models.User
	err := r.pool.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, apperrors.ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("select user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetRoleByID(ctx context.Context, userID int64) (string, error) {
	var role string
	err := r.pool.QueryRow(
		ctx,
		`SELECT role
		FROM app_user
		WHERE user_id = $1`,
		userID,
	).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("get user role: %w", err)
	}
	return role, nil
}

func (r *PostgresRepository) HasEmployerProfile(ctx context.Context, userID int64) (bool, error) {
	var exist bool
	err := r.pool.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM employer_profile
			WHERE user_id = $1
		)`,
		userID,
	).Scan(&exist)

	if err != nil {
		return false, fmt.Errorf("check employer profile: %w", err)
	}

	return exist, nil
}
