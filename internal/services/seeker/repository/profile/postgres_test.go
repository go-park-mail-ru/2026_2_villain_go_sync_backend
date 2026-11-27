package profile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/repository/profile"
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
		"city", "messenger", "contact_email", "preferred_position", "preferred_cities", "work_format",
	}).AddRow(1, "Ivan", "Ivanov", "12345", "About me", nil, now, now, nil, nil, nil, nil, nil, nil)

	mockPool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	mockPool.ExpectQuery(`SELECT user_id`).
		WithArgs(1).
		WillReturnRows(rows)
	expectCareerDetails(mockPool, "seeker")
	mockPool.ExpectCommit()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

	p, err := repo.GetSeekerByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.PersonalInfo.FirstName != "Ivan" {
		t.Errorf("expected Ivan, got %s", p.PersonalInfo.FirstName)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPostgresRepository_GetSeekerByID_NotFound(t *testing.T) {
	mockPool, _ := pgxmock.NewPool()
	defer mockPool.Close()

	repo := profile.NewPostgresRepository(mockPool)

	mockPool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	mockPool.ExpectQuery(`SELECT user_id`).
		WithArgs(1).
		WillReturnError(pgx.ErrNoRows)
	mockPool.ExpectRollback()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

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

	mockPool.ExpectBegin()
	mockPool.ExpectQuery(`INSERT INTO seeker_profile`).
		WithArgs(1, "Ivan", "Ivanov", "12345", "About me", (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), (*models.WorkFormat)(nil)).
		WillReturnRows(
			pgxmock.NewRows([]string{"created_at", "updated_at", "avatar_key", "city", "messenger", "contact_email", "preferred_position", "preferred_cities", "work_format"}).
				AddRow(now, now, nil, nil, nil, nil, nil, nil, nil),
		)
	expectCareerDetails(mockPool, "seeker")
	mockPool.ExpectCommit()
	mockPool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

	p := models.SeekerProfile{
		Profile:      models.Profile{UserID: 1, Description: "About me"},
		PersonalInfo: models.SeekerPersonalInfo{FirstName: "Ivan", LastName: "Ivanov"},
		Contacts:     models.SeekerContacts{Phone: "12345"},
	}

	res, err := repo.UpsertSeeker(context.Background(), p)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Profile.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", res.Profile.UserID)
	}
	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func expectCareerDetails(pool pgxmock.PgxPoolIface, prefix string) {
	for _, section := range []string{"experience", "education", "skill", "project"} {
		pool.ExpectQuery("FROM " + prefix + "_" + section).
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows(nil))
	}
}

func TestPostgresRepository_UpdateSeekerAvatarKey(t *testing.T) {
	dbErr := errors.New("update failed")
	for _, tc := range []struct {
		name    string
		rows    int64
		execErr error
		wantErr error
	}{
		{name: "success", rows: 1},
		{name: "missing seeker", wantErr: apperrors.ErrNotFound},
		{name: "database error", execErr: dbErr, wantErr: dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := pgxmock.NewPool()
			if err != nil {
				t.Fatal(err)
			}
			defer pool.Close()

			key := "avatars/42/avatar"
			exec := pool.ExpectExec("UPDATE seeker_profile").WithArgs(&key, 42)
			if tc.execErr != nil {
				exec.WillReturnError(tc.execErr)
			} else {
				exec.WillReturnResult(pgxmock.NewResult("UPDATE", tc.rows))
			}
			err = profile.NewPostgresRepository(pool).UpdateSeekerAvatarKey(t.Context(), 42, &key)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("UpdateSeekerAvatarKey() = %v, want %v", err, tc.wantErr)
			}
			if err := pool.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
