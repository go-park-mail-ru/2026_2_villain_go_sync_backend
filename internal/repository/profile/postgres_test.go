package profile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/profile"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func TestPostgresRepository_GetSeekerByID_Success(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mockPool.Close()

	repo := profile.NewPostgresRepository(mockPool)

	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"user_id", "first_name", "last_name", "phone",
		"about", "avatar_key", "created_at", "updated_at",
	}).AddRow(1, "Ivan", "Ivanov", "12345", "About me", nil, now, now)

	mockPool.ExpectQuery(`SELECT user_id, first_name`).
		WithArgs(1).
		WillReturnRows(rows)

	p, err := repo.GetSeekerByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.FirstName != "Ivan" {
		t.Errorf("expected Ivan, got %s", p.FirstName)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetSeekerByID_NotFound(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := profile.NewPostgresRepository(mockPool)

	mockPool.ExpectQuery(`SELECT user_id`).
		WithArgs(1).
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.GetSeekerByID(context.Background(), 1)

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresRepository_UpsertSeeker_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := profile.NewPostgresRepository(mockPool)

	now := time.Now()

	mockPool.ExpectQuery(`INSERT INTO seeker_profile`).
		WithArgs(1, "Ivan", "Ivanov", "12345", "About me").
		WillReturnRows(
			pgxmock.NewRows([]string{"created_at", "updated_at", "avatar_key"}).
				AddRow(now, now, nil),
		)

	p := models.SeekerProfile{
		Profile:   models.Profile{UserID: 1, Description: "About me"},
		FirstName: "Ivan",
		LastName:  "Ivanov",
		Phone:     "12345",
	}

	res, err := repo.UpsertSeeker(context.Background(), p)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", res.UserID)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
