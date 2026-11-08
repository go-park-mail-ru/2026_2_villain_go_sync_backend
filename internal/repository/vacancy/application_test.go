package vacancy_test

import (
	"context"
	"testing"

	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
	"github.com/pashagolub/pgxmock/v5"
)

func TestPostgresRepository_IsOwner_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := vacancy.NewPostgresRepository(mockPool)

	rows := pgxmock.NewRows([]string{"exists"}).AddRow(true)

	mockPool.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(1), int64(10)).
		WillReturnRows(rows)

	owned, err := repo.IsOwner(context.Background(), 1, 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !owned {
		t.Error("expected true, got false")
	}
}

func TestPostgresRepository_ListByVacancyID_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := vacancy.NewPostgresRepository(mockPool)
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"first_name", "last_name", "status",
		"applied_at", "resume_title",
		"application_id", "vacancy_id", "vacancy_title",
		"seeker_id", "resume_id", "cover_letter",
	}).AddRow(
		"Ivan", "Ivanov", "created",
		now, "Go dev",
		int64(5), int64(1), "Go dev",
		int64(12), int64(7), nil,
	)

	mockPool.ExpectQuery(`SELECT`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	applicants, err := repo.ListByVacancyID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(applicants) != 1 {
		t.Errorf("expected 1 applicant, got %d", len(applicants))
	}

	if len(applicants) == 1 {
		if applicants[0].ApplicationID != 5 {
			t.Errorf("unexpected application ID: %d", applicants[0].ApplicationID)
		}
		if applicants[0].ResumeID != 7 {
			t.Errorf("unexpected resume ID: %d", applicants[0].ResumeID)
		}
	}
}
