package vacancy_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
	"github.com/jackc/pgx/v5"
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

	wantCategories := []string{"Backend", "Go"}
	expectVacancyCategories(mockPool, wantCategories)
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
	if !slices.Equal(res.Categories, wantCategories) {
		t.Errorf("categories = %v, want %v", res.Categories, wantCategories)
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
		"work_format", "work_schedule", "work_hours", "city", "office_address", "categories",
	}).AddRow(
		1, 1, 10, "Test Company",
		"Go dev", "desc", nil, nil, models.CurrencyRUB, now, "", "", "",
		now, now, nil, nil, nil, nil, nil, nil, nil, nil, nil, []string{"Backend"},
	)

	mockPool.ExpectQuery(`SELECT`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	res, err := repo.GetByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(res.Categories, []string{"Backend"}) {
		t.Errorf("unexpected categories: %v", res.Categories)
	}
	if res.Title != "Go dev" {
		t.Errorf("expected Go dev, got %s", res.Title)
	}
	if res.Company.LogoURL != nil {
		t.Errorf("expected nil logo URL, got %v", res.Company.LogoURL)
	}
}

func expectVacancyCategories(mockPool pgxmock.PgxPoolIface, names []string) {
	mockPool.ExpectQuery(`SELECT ARRAY\(.*FROM vacancy_category`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"categories"}).AddRow(names))
}

func TestPostgresRepository_Update_ReturnsCategoryNames(t *testing.T) {
	for _, tc := range []struct {
		name string
		ids  *[]int64
		want []string
	}{
		{name: "unchanged", want: []string{"Backend"}},
		{name: "replaced", ids: &[]int64{2}, want: []string{"Go"}},
		{name: "cleared", ids: &[]int64{}, want: []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			if err != nil {
				t.Fatal(err)
			}
			defer mockPool.Close()

			repo := vacancy.NewPostgresRepository(mockPool)
			now := time.Now()
			mockPool.ExpectBegin()
			mockPool.ExpectQuery("SELECT employer_id").WithArgs(int64(1)).
				WillReturnRows(pgxmock.NewRows([]string{"employer_id"}).AddRow(int64(10)))
			mockPool.ExpectQuery(`SELECT COALESCE\(MAX\(version\)`).
				WithArgs(int64(1)).
				WillReturnRows(pgxmock.NewRows([]string{"version"}).AddRow(1))
			mockPool.ExpectQuery("WITH new_version AS").
				WithArgs(int64(1), 2, "Go dev", "desc", (*float64)(nil), (*float64)(nil), models.CurrencyRUB, "Develop services", "Know Go",
					(*string)(nil), (*models.ExperienceLevel)(nil), (*models.EmploymentType)(nil), (*models.WorkFormat)(nil),
					(*models.WorkSchedule)(nil), (*string)(nil), (*string)(nil), (*string)(nil)).
				WillReturnRows(pgxmock.NewRows([]string{
					"changed_at", "offer", "experience_level", "employment_type", "work_format",
					"work_schedule", "work_hours", "city", "office_address",
				}).AddRow(now, nil, nil, nil, nil, nil, nil, nil, nil))
			if tc.ids != nil {
				mockPool.ExpectExec("DELETE FROM vacancy_category").WithArgs(int64(1)).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				if len(*tc.ids) > 0 {
					mockPool.ExpectExec("INSERT INTO vacancy_category").WithArgs(int64(1), *tc.ids).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
				}
			}
			mockPool.ExpectQuery("UPDATE vacancy SET updated_at").WithArgs(int64(1)).
				WillReturnRows(pgxmock.NewRows([]string{"updated_at", "apply_until"}).AddRow(now, nil))
			expectVacancyCategories(mockPool, tc.want)
			mockPool.ExpectCommit()

			got, err := repo.Update(t.Context(), models.Vacancy{
				ID:           1,
				Version:      1,
				Title:        "Go dev",
				Company:      models.VacancyCompany{EmployerID: 10},
				Content:      models.VacancyContent{Description: "desc", Responsibilities: "Develop services"},
				Requirements: models.VacancyRequirements{Description: "Know Go"},
				Salary:       models.VacancySalary{Currency: models.CurrencyRUB},
				Categories:   []string{"Stale"},
			}, tc.ids)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got.Categories, tc.want) || got.Categories == nil {
				t.Errorf("categories = %#v, want %#v", got.Categories, tc.want)
			}
			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPostgresRepository_ListByEmployer_EmptyPage(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPool.Close()

	mockPool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	mockPool.ExpectQuery(`SELECT COUNT\(\*\).*AND v.employer_id = \$1`).
		WithArgs(int64(10)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(6)))
	mockPool.ExpectQuery(`AND v.employer_id = \$1 ORDER BY v.created_at ASC, v.vacancy_id ASC LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(10), 5, int64(10)).
		WillReturnRows(pgxmock.NewRows(nil))
	mockPool.ExpectCommit()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

	items, total, err := vacancy.NewPostgresRepository(mockPool).ListByEmployer(context.Background(), 10, 3, 5)
	if err != nil || total != 6 || items == nil || len(items) != 0 {
		t.Fatalf("items = %v, total = %d, err = %v", items, total, err)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
