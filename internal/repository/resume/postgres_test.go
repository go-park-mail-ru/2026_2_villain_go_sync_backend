package resume_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/resume"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func TestPostgresRepository_Create_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{"resume_id", "created_at", "updated_at"}).
		AddRow(1, now, now)

	mockPool.ExpectQuery(`INSERT INTO resume`).
		WithArgs(10, "Go dev", "desc", "exp", "edu").
		WillReturnRows(rows)

	res, err := repo.Create(context.Background(), models.Resume{
		SeekerID:    10,
		Title:       "Go dev",
		Description: "desc",
		Experience:  "exp",
		Education:   "edu",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 {
		t.Errorf("expected ID 1, got %d", res.ID)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetByID_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{"resume_id", "seeker_id", "title", "description", "experience", "education", "created_at", "updated_at"}).
		AddRow(1, 10, "Go dev", "desc", "exp", "edu", now, now)

	mockPool.ExpectQuery(`SELECT resume_id`).
		WithArgs(1).
		WillReturnRows(rows)

	res, err := repo.GetByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "Go dev" {
		t.Errorf("expected Go dev, got %s", res.Title)
	}
}

func TestPostgresRepository_GetByID_NotFound(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)

	mockPool.ExpectQuery(`SELECT resume_id`).
		WithArgs(1).
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 1)

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresRepository_ListBySeekerID_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{"resume_id", "seeker_id", "title", "description", "experience", "education", "created_at", "updated_at"}).
		AddRow(1, 10, "Go dev", "desc", "exp", "edu", now, now).
		AddRow(2, 10, "C++ dev", "desc", "exp", "edu", now, now)

	mockPool.ExpectQuery(`SELECT resume_id`).
		WithArgs(10).
		WillReturnRows(rows)

	resumes, err := repo.ListBySeekerID(context.Background(), 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resumes) != 2 {
		t.Errorf("expected 2 resumes, got %d", len(resumes))
	}
}

func TestPostgresRepository_Update_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"seeker_id", "created_at", "updated_at",
	}).AddRow(10, now, now)

	mockPool.ExpectQuery(`UPDATE resume`).
		WithArgs("Updated Title", "desc", "exp", "edu", 1, 10).
		WillReturnRows(rows)

	res, err := repo.Update(context.Background(), models.Resume{
		ID:          1,
		Title:       "Updated Title",
		Description: "desc",
		Experience:  "exp",
		Education:   "edu",
	}, 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
	if res.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPostgresRepository_Delete_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)

	mockPool.ExpectExec(`DELETE FROM resume`).
		WithArgs(1, 10).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(context.Background(), 1, 10)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostgresRepository_Delete_NoContent(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)

	mockPool.ExpectExec(`DELETE FROM resume`).
		WithArgs(1, 10).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.Delete(context.Background(), 1, 10)

	if !errors.Is(err, apperrors.ErrNoContent) {
		t.Errorf("expected ErrNoContent, got %v", err)
	}
}
