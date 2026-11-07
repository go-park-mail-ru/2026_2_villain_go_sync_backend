package vacancy_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
	"github.com/pashagolub/pgxmock/v5"
)

func TestPostgresRepository_Create_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := vacancy.NewPostgresRepository(mockPool)
	now := time.Now()

	mockPool.ExpectBegin()

	rows := pgxmock.NewRows([]string{"vacancy_id", "created_at"}).AddRow(1, now)
	mockPool.ExpectQuery(`INSERT INTO vacancy`).
		WithArgs(int64(10)).
		WillReturnRows(rows)

	mockPool.ExpectExec(`INSERT INTO vacancy_history`).
		WithArgs(int64(1), "Go dev", "desc", (*float64)(nil), (*float64)(nil), "RUB").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockPool.ExpectExec(`INSERT INTO vacancy_category`).
		WithArgs(int64(1), []int64{1, 2}).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))

	mockPool.ExpectCommit()

	v := models.Vacancy{
		EmployerID:  10,
		Title:       "Go dev",
		Description: "desc",
		Currency:    "RUB",
	}

	res, err := repo.Create(context.Background(), v, []int64{1, 2})

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

	repo := vacancy.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{"vacancy_id", "version", "employer_id", "title", "description", "salary_from", "salary_to", "currency", "created_at"}).
		AddRow(1, 1, 10, "Go dev", "desc", nil, nil, "RUB", now)

	mockPool.ExpectQuery(`SELECT`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	res, err := repo.GetByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "Go dev" {
		t.Errorf("expected Go dev, got %s", res.Title)
	}
}
