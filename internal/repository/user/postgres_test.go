package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

func TestPostgresRepository_Create_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := user.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{"user_id", "created_at"}).AddRow(1, now)

	mockPool.ExpectQuery(`INSERT INTO app_user`).
		WithArgs("test@example.com", "hash", models.RoleSeeker).
		WillReturnRows(rows)

	usr, err := repo.Create(context.Background(), models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleSeeker,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usr.ID != 1 {
		t.Errorf("expected ID 1, got %d", usr.ID)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_Create_EmailTaken(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := user.NewPostgresRepository(mockPool)

	pgErr := &pgconn.PgError{Code: "23505"}

	mockPool.ExpectQuery(`INSERT INTO app_user`).
		WithArgs("test@example.com", "hash", models.RoleSeeker).
		WillReturnError(pgErr)

	_, err := repo.Create(context.Background(), models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         models.RoleSeeker,
	})

	if !errors.Is(err, apperrors.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestPostgresRepository_GetByEmail_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := user.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{"user_id", "email", "password_hash", "role", "created_at"}).
		AddRow(1, "test@example.com", "hash", models.RoleSeeker, now)

	mockPool.ExpectQuery(`SELECT user_id, email, password_hash, role, created_at FROM app_user`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	usr, err := repo.GetByEmail(context.Background(), "test@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usr.Email != "test@example.com" {
		t.Errorf("expected test@example.com, got %s", usr.Email)
	}
}

func TestPostgresRepository_GetByEmail_NotFound(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := user.NewPostgresRepository(mockPool)

	mockPool.ExpectQuery(`SELECT user_id`).
		WithArgs("missing@example.com").
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.GetByEmail(context.Background(), "missing@example.com")

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresRepository_GetRoleByID_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := user.NewPostgresRepository(mockPool)

	rows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEmployer)

	mockPool.ExpectQuery(`SELECT role FROM app_user`).
		WithArgs(int64(10)).
		WillReturnRows(rows)

	role, err := repo.GetRoleByID(context.Background(), 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role != string(models.RoleEmployer) {
		t.Errorf("expected employer, got %s", role)
	}
}

func TestPostgresRepository_HasEmployerProfile_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := user.NewPostgresRepository(mockPool)

	rows := pgxmock.NewRows([]string{"exists"}).AddRow(true)

	mockPool.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(10)).
		WillReturnRows(rows)

	hasProfile, err := repo.HasEmployerProfile(context.Background(), 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasProfile {
		t.Error("expected true, got false")
	}
}
