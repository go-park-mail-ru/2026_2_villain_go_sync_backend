package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/application"
	"github.com/pashagolub/pgxmock/v5"
)

func TestPostgresRepository_Create_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := application.NewPostgresRepository(mockPool)
	now := time.Now()

	mockPool.ExpectBegin()

	coverLetter := "Cover letter"
	rows := pgxmock.NewRows([]string{"application_id", "created_at", "updated_at"}).AddRow(1, now, now)
	mockPool.ExpectQuery(`FROM create_application`).
		WithArgs(int64(10), 1, int64(5), int64(2), &coverLetter).
		WillReturnRows(rows)

	mockPool.ExpectCommit()

	app := models.Application{
		SeekerID:       2,
		VacancyID:      10,
		VacancyVersion: 1,
		ResumeID:       5,
		CoverLetter:    &coverLetter,
	}
	res, err := repo.Create(context.Background(), app)

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
