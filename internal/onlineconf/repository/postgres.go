package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type database interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PgxPool interface {
	database
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

type Store struct {
	Scopes    ScopeRepository
	Configs   ConfigRepository
	Flags     FeatureFlagRepository
	Overrides FeatureFlagOverrideRepository
	Admins    AdminRepository
	History   HistoryRepository
	Keys      KeyRepository
	Snapshots SnapshotRepository

	conn connection
}

type connection struct {
	db   database
	pool PgxPool
	tx   pgx.Tx
}

func NewStore(pool PgxPool) *Store {
	return newStore(connection{db: pool, pool: pool})
}

func newStore(conn connection) *Store {
	return &Store{
		Scopes:    ScopeRepository{db: conn.db},
		Configs:   ConfigRepository{conn: conn},
		Flags:     FeatureFlagRepository{conn: conn},
		Overrides: FeatureFlagOverrideRepository{db: conn.db},
		Admins:    AdminRepository{db: conn.db},
		History:   HistoryRepository{db: conn.db},
		Keys:      KeyRepository{db: conn.db},
		Snapshots: SnapshotRepository{db: conn.db},
		conn:      conn,
	}
}

func (s *Store) WithinTx(ctx context.Context, fn func(*Store) error) error {
	return s.conn.withinTx(ctx, func(tx pgx.Tx) error {
		return fn(newStore(connection{db: tx, pool: s.conn.pool, tx: tx}))
	})
}

func (c connection) withinTx(ctx context.Context, fn func(pgx.Tx) error) error {
	if c.tx != nil {
		return classifyError(ctx, pgx.BeginFunc(ctx, c.tx, fn))
	}
	return classifyError(ctx, pgx.BeginTxFunc(ctx, c.pool, pgx.TxOptions{IsoLevel: pgx.RepeatableRead}, fn))
}

var ErrRetryable = errors.New("retryable database transaction")

func classifyError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{apperrors.ErrNotFound, apperrors.ErrConflict, apperrors.ErrInvalidInput, ErrRetryable} {
		if errors.Is(err, known) {
			return err
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %w", apperrors.ErrNotFound, err)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23503", "23P01":
			return fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		case "23502", "23514", "22001", "22003", "22P02", "22P05":
			return fmt.Errorf("%w: %w", apperrors.ErrInvalidInput, err)
		case "40001", "40P01":
			return fmt.Errorf("%w: %w", ErrRetryable, err)
		case "57014":
			if ctx.Err() != nil {
				return fmt.Errorf("%w: %w", ctx.Err(), err)
			}
		}
	}
	return err
}

func execOne(ctx context.Context, db database, sql string, args ...any) error {
	tag, err := db.Exec(ctx, sql, args...)
	if err != nil {
		return classifyError(ctx, err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
