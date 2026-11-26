package vacancy_test

import (
	"context"
	"testing"

	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/repository/vacancy"
	"github.com/jackc/pgx/v5"
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

func TestPostgresRepository_ApplicantAvatarURL(t *testing.T) {
	tests := []struct {
		name      string
		byVacancy bool
		avatarKey string
		wantURL   string
	}{
		{
			name:      "vacancy with avatar",
			byVacancy: true,
			avatarKey: "avatars/12/photo",
			wantURL:   "/api/seekers/12/avatar",
		},
		{
			name:      "vacancy without avatar",
			byVacancy: true,
		},
		{
			name:      "employer with avatar",
			avatarKey: "avatars/12/photo",
			wantURL:   "/api/seekers/12/avatar",
		},
		{
			name: "employer without avatar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			if err != nil {
				t.Fatal(err)
			}
			defer mockPool.Close()

			repo := vacancy.NewPostgresRepository(mockPool)
			now := time.Now()
			resumeID := int64(7)

			rows := pgxmock.NewRows([]string{
				"first_name", "last_name", "status",
				"applied_at", "resume_title",
				"application_id", "vacancy_id", "vacancy_title",
				"seeker_id", "resume_id", "cover_letter",
				"avatar_key", "updated_at", "vacancy_version", "status_history_id",
				"status_comment", "status_changed_at", "status_changed_by", "reason_code",
			}).AddRow(
				"Ivan", "Ivanov", models.ApplicationCreated,
				now, "Go dev",
				int64(5), int64(1), "Go dev",
				int64(12), &resumeID, nil,
				tt.avatarKey, now, 2, nil, nil, nil, nil, nil,
			)

			filter := `v.employer_id = \$1`
			if tt.byVacancy {
				filter = `a.vacancy_id = \$1`
			}
			mockPool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
			mockPool.ExpectQuery(`SELECT COUNT\(\*\).*WHERE ` + filter).
				WithArgs(int64(1)).
				WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(8)))
			mockPool.ExpectQuery(`WHERE `+filter+` AND \(a.created_at, a.application_id\) < \(\$2, \$3\) ORDER BY a.created_at DESC, a.application_id DESC LIMIT \$4`).
				WithArgs(int64(1), now, int64(9), 4).
				WillReturnRows(rows)
			mockPool.ExpectCommit()
			mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

			var applicants []models.Applicant
			var total int64

			if tt.byVacancy {
				applicants, total, err = repo.ListByVacancyID(
					context.Background(), 1, pagination.Params{Cursor: &pagination.Cursor{CreatedAt: now, ID: 9}, Limit: 4},
				)
			} else {
				applicants, total, err = repo.ListByEmployerID(
					context.Background(), 1, pagination.Params{Cursor: &pagination.Cursor{CreatedAt: now, ID: 9}, Limit: 4},
				)
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if total != 8 {
				t.Fatalf("total = %d, want 8", total)
			}

			if len(applicants) != 1 {
				t.Fatalf("expected 1 applicant, got %d", len(applicants))
			}

			app := applicants[0]

			if app.ApplicationID != 5 || app.Vacancy.VacancyVersion != 2 || !app.CreatedAt.Equal(now) || !app.UpdatedAt.Equal(now) || app.Resume.ResumeID == nil || *app.Resume.ResumeID != 7 {
				t.Fatalf("unexpected applicant IDs: %+v", app)
			}

			if tt.wantURL == "" {
				if app.Seeker.AvatarURL != nil {
					t.Errorf("expected nil avatar URL, got %q", *app.Seeker.AvatarURL)
				}
			} else {
				if app.Seeker.AvatarURL == nil {
					t.Fatal("expected avatar URL, got nil")
				}
				if *app.Seeker.AvatarURL != tt.wantURL {
					t.Errorf("got %q, want %q", *app.Seeker.AvatarURL, tt.wantURL)
				}
			}

			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
