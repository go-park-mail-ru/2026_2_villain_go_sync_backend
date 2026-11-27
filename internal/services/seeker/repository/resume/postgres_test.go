package resume_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/repository/resume"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func TestPostgresRepository_Create_Success(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := resume.NewPostgresRepository(mockPool)
	now := time.Now()
	targetPosition := "Engineer"

	rows := pgxmock.NewRows([]string{"resume_id", "created_at", "updated_at", "target_position", "city", "employment_type", "work_format", "work_schedule", "salary_from", "portfolio_url", "languages"}).
		AddRow(1, now, now, &targetPosition, nil, nil, nil, nil, nil, nil, nil)

	mockPool.ExpectBegin()
	mockPool.ExpectQuery(`INSERT INTO resume`).
		WithArgs(10, "Go dev", "desc", "exp", "edu", &targetPosition, (*string)(nil), (*models.EmploymentType)(nil), (*models.WorkFormat)(nil), (*models.WorkSchedule)(nil), (*float64)(nil), (*string)(nil), (*string)(nil)).
		WillReturnRows(rows)
	expectCareerDetails(mockPool, "resume")
	mockPool.ExpectCommit()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

	res, err := repo.Create(context.Background(), models.Resume{
		SeekerID:    10,
		Title:       "Go dev",
		Preferences: models.ResumePreferences{TargetPosition: &targetPosition},
		Content: models.ResumeContent{
			Description: "desc",
			Experience:  "exp",
			Education:   "edu",
		},
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

	rows := pgxmock.NewRows([]string{"resume_id", "seeker_id", "title", "description", "experience", "education", "created_at", "updated_at", "target_position", "city", "employment_type", "work_format", "work_schedule", "salary_from", "portfolio_url", "languages"}).
		AddRow(1, 10, "Go dev", "desc", "exp", "edu", now, now, nil, nil, nil, nil, nil, nil, nil, nil)

	mockPool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	mockPool.ExpectQuery(`SELECT resume_id`).
		WithArgs(1).
		WillReturnRows(rows)
	expectCareerDetails(mockPool, "resume")
	mockPool.ExpectCommit()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

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

	mockPool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	mockPool.ExpectQuery(`SELECT resume_id`).
		WithArgs(1).
		WillReturnError(pgx.ErrNoRows)
	mockPool.ExpectRollback()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

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

	rows := pgxmock.NewRows([]string{"resume_id", "seeker_id", "title", "description", "experience", "education", "created_at", "updated_at", "target_position", "city", "employment_type", "work_format", "work_schedule", "salary_from", "portfolio_url", "languages"}).
		AddRow(1, 10, "Go dev", "desc", "exp", "edu", now, now, nil, nil, nil, nil, nil, nil, nil, nil).
		AddRow(2, 10, "C++ dev", "desc", "exp", "edu", now, now, nil, nil, nil, nil, nil, nil, nil, nil)

	mockPool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	mockPool.ExpectQuery(`SELECT resume_id`).
		WithArgs(10).
		WillReturnRows(rows)
	expectCareerDetails(mockPool, "resume")
	mockPool.ExpectCommit()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

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
	targetPosition := "Engineer"

	rows := pgxmock.NewRows([]string{
		"seeker_id", "created_at", "updated_at",
		"target_position", "city", "employment_type", "work_format", "work_schedule", "salary_from", "portfolio_url", "languages",
	}).AddRow(10, now, now, &targetPosition, nil, nil, nil, nil, nil, nil, nil)

	mockPool.ExpectBegin()
	mockPool.ExpectQuery(`UPDATE resume`).
		WithArgs("Updated Title", "desc", "exp", "edu", 1, 10, &targetPosition, (*string)(nil), (*models.EmploymentType)(nil), (*models.WorkFormat)(nil), (*models.WorkSchedule)(nil), (*float64)(nil), (*string)(nil), (*string)(nil), true).
		WillReturnRows(rows)
	expectCareerDetails(mockPool, "resume")
	mockPool.ExpectCommit()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

	res, err := repo.Update(context.Background(), dto.UpdateInput{Resume: models.Resume{
		ID:          1,
		Title:       "Updated Title",
		Preferences: models.ResumePreferences{TargetPosition: &targetPosition},
		Content: models.ResumeContent{
			Description: "desc",
			Experience:  "exp",
			Education:   "edu",
		},
	}, ClearSalaryFrom: true}, 10)

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

func expectCareerDetails(pool pgxmock.PgxPoolIface, prefix string) {
	for _, section := range []string{"experience", "education", "skill", "project"} {
		pool.ExpectQuery("FROM " + prefix + "_" + section).
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows(nil))
	}
}

func TestPostgresRepository_HasApplicantResume(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	pool.ExpectQuery(`SELECT EXISTS\(.*FROM application a JOIN vacancy v`).
		WithArgs(7, 42).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	allowed, err := resume.NewPostgresRepository(pool).HasApplicantResume(t.Context(), 7, 42)
	if err != nil || !allowed {
		t.Fatalf("HasApplicantResume() = %t, %v, want true, nil", allowed, err)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
