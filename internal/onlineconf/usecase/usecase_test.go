package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func TestScopeRollsBackWhenHistoryWriteFails(t *testing.T) {
	pool, err := pgxmock.NewPool()
	requireError(t, err, nil)
	defer pool.Close()

	const actor = int64(1)
	now := time.Now()
	pool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	pool.ExpectQuery("FROM public.admin WHERE user_id = .* FOR SHARE").WithArgs(actor).
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "role", "granted_by", "created_at"}).
			AddRow(actor, "admin", nil, now))
	pool.ExpectQuery("INSERT INTO public.scope").WithArgs("scope", (*string)(nil)).
		WillReturnRows(pgxmock.NewRows([]string{"scope_id", "name", "description", "created_at", "updated_at"}).
			AddRow(int64(10), "scope", nil, now, now))
	historyErr := errors.New("history insert failed")
	pool.ExpectQuery("INSERT INTO public.change_history").
		WithArgs(actor, models.ResourceTypeScope, int64(10), pgxmock.AnyArg(), "scope", models.ActionCreate, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(historyErr)
	pool.ExpectRollback()
	pool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

	got, err := usecase.New(repository.NewStore(pool)).Scopes.Create(t.Context(), actor, models.Scope{Name: "scope"})
	requireError(t, err, historyErr)
	if got.ID != 0 {
		t.Fatalf("Create returned uncommitted scope: %+v", got)
	}
	requireError(t, pool.ExpectationsWereMet(), nil)
}

func TestUnauthorizedOperationStopsBeforeResourceLookup(t *testing.T) {
	pool, err := pgxmock.NewPool()
	requireError(t, err, nil)
	defer pool.Close()

	pool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	pool.ExpectQuery("FROM public.admin WHERE user_id = .* FOR SHARE").WithArgs(int64(88)).WillReturnError(pgx.ErrNoRows)
	pool.ExpectRollback()
	pool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)
	_, err = usecase.New(repository.NewStore(pool)).Configs.Get(t.Context(), 88, 42)
	requireError(t, err, apperrors.ErrForbidden)
	requireError(t, pool.ExpectationsWereMet(), nil)
}

func requireError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("error = %v, want %v", got, want)
	}
}
