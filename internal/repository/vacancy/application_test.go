package vacancy_test

import (
	"context"
	"testing"

	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
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

			rows := pgxmock.NewRows([]string{
				"first_name", "last_name", "status",
				"applied_at", "resume_title",
				"application_id", "vacancy_id", "vacancy_title",
				"seeker_id", "resume_id", "cover_letter",
				"avatar_key",
			}).AddRow(
				"Ivan", "Ivanov", "created",
				now, "Go dev",
				int64(5), int64(1), "Go dev",
				int64(12), int64(7), nil,
				tt.avatarKey,
			)

			mockPool.ExpectQuery(`FROM application a`).
				WithArgs(int64(1)).
				WillReturnRows(rows)

			var applicants []models.Applicant

			if tt.byVacancy {
				applicants, err = repo.ListByVacancyID(
					context.Background(), 1,
				)
			} else {
				applicants, err = repo.ListByEmployerID(
					context.Background(), 1,
				)
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(applicants) != 1 {
				t.Fatalf("expected 1 applicant, got %d", len(applicants))
			}

			app := applicants[0]

			if app.ApplicationID != 5 || app.ResumeID != 7 {
				t.Fatalf("unexpected applicant IDs: %+v", app)
			}

			if tt.wantURL == "" {
				if app.AvatarURL != nil {
					t.Errorf("expected nil avatar URL, got %q", *app.AvatarURL)
				}
			} else {
				if app.AvatarURL == nil {
					t.Fatal("expected avatar URL, got nil")
				}
				if *app.AvatarURL != tt.wantURL {
					t.Errorf("got %q, want %q", *app.AvatarURL, tt.wantURL)
				}
			}

			if err := mockPool.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
