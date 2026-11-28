package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
)

func TestClassifyError(t *testing.T) {
	for _, tc := range []struct {
		cause error
		want  error
	}{
		{pgx.ErrNoRows, apperrors.ErrNotFound},
		{&pgconn.PgError{Code: "23505"}, apperrors.ErrConflict},
		{&pgconn.PgError{Code: "23503"}, apperrors.ErrConflict},
		{&pgconn.PgError{Code: "23514"}, apperrors.ErrInvalidInput},
		{&pgconn.PgError{Code: "23502"}, apperrors.ErrInvalidInput},
		{&pgconn.PgError{Code: "40001"}, apperrors.ErrRetryable},
		{&pgconn.PgError{Code: "40P01"}, apperrors.ErrRetryable},
		{context.DeadlineExceeded, context.DeadlineExceeded},
	} {
		got := classifyError(t.Context(), tc.cause)
		if !errors.Is(got, tc.want) || !errors.Is(got, tc.cause) {
			t.Errorf("classifyError(%v) = %v, want %v and original error", tc.cause, got, tc.want)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := classifyError(ctx, &pgconn.PgError{Code: "57014"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("query cancellation = %v", err)
	}
}

func TestWithinTxRollbackAndCommitError(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := NewStore(pool)
	callbackErr := errors.New("audit failed")
	pool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	pool.ExpectRollback()
	pool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)
	if err := store.WithinTx(t.Context(), func(*Store) error { return callbackErr }); !errors.Is(err, callbackErr) {
		t.Fatalf("WithinTx rollback = %v", err)
	}

	pool.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	pool.ExpectCommit().WillReturnError(&pgconn.PgError{Code: "23514"})
	pool.ExpectRollback().WillReturnError(pgx.ErrTxClosed)
	if err := store.WithinTx(t.Context(), func(*Store) error { return nil }); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("WithinTx deferred constraint = %v", err)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
