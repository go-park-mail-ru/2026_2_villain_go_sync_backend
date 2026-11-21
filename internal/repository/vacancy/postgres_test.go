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
	employment := models.EmploymentFull

	mockPool.ExpectBegin()

	rows := pgxmock.NewRows([]string{"vacancy_id", "created_at", "updated_at", "apply_until"}).AddRow(1, now, now, nil)
	mockPool.ExpectQuery(`INSERT INTO vacancy`).
		WithArgs(int64(10)).
		WillReturnRows(rows)

	mockPool.ExpectQuery(`INSERT INTO vacancy_history`).
		WithArgs(int64(1), "Go dev", "desc", (*float64)(nil), (*float64)(nil), models.CurrencyRUB, "Develop services", "Know Go",
			(*string)(nil), (*models.ExperienceLevel)(nil), &employment, (*models.WorkFormat)(nil),
			(*models.WorkSchedule)(nil), (*string)(nil), (*string)(nil), (*string)(nil)).
		WillReturnRows(pgxmock.NewRows([]string{
			"changed_at", "offer", "experience_level", "employment_type", "work_format",
			"work_schedule", "work_hours", "city", "office_address",
		}).AddRow(now, nil, nil, &employment, nil, nil, nil, nil, nil))

	mockPool.ExpectExec(`INSERT INTO vacancy_category`).
		WithArgs(int64(1), []int64{1, 2}).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))

	mockPool.ExpectQuery(`SELECT company_name`).
		WithArgs(int64(10)).
		WillReturnRows(
			pgxmock.NewRows([]string{"company_name", "logo_key"}).
				AddRow("Test Company", "logos/10/test"),
		)

	mockPool.ExpectCommit()

	v := models.Vacancy{
		Title:      "Go dev",
		Conditions: models.VacancyConditions{EmploymentType: &employment},
		Company: models.VacancyCompany{
			EmployerID: 10,
		},
		Content: models.VacancyContent{
			Description:      "desc",
			Responsibilities: "Develop services",
		},
		Requirements: models.VacancyRequirements{
			Description: "Know Go",
		},
		Salary: models.VacancySalary{
			Currency: models.CurrencyRUB,
		},
	}

	res, err := repo.Create(context.Background(), v, []int64{1, 2})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 {
		t.Errorf("expected ID 1, got %d", res.ID)
	}

	if res.Company.Name != "Test Company" {
		t.Errorf("expected Test Company, got %s", res.Company.Name)
	}

	if res.Company.LogoURL == nil || *res.Company.LogoURL != "/api/employers/10/logo" {
		t.Errorf("unexpected logo URL: %v", res.Company.LogoURL)
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

	rows := pgxmock.NewRows([]string{
		"vacancy_id", "version", "employer_id", "company_name",
		"title", "description", "salary_from", "salary_to",
		"currency", "created_at", "logo_key", "responsibilities", "requirements",
		"updated_at", "changed_at", "apply_until", "offer", "experience_level", "employment_type",
		"work_format", "work_schedule", "work_hours", "city", "office_address",
	}).AddRow(
		1, 1, 10, "Test Company",
		"Go dev", "desc", nil, nil, models.CurrencyRUB, now, "", "", "",
		now, now, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

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
	if res.Company.LogoURL != nil {
		t.Errorf("expected nil logo URL, got %v", res.Company.LogoURL)
	}
}
