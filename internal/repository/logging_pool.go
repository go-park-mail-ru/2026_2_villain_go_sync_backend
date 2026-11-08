package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LoggingPool struct {
	next *pgxpool.Pool
}

func NewLoggingPool(next *pgxpool.Pool) *LoggingPool {
	return &LoggingPool{next: next}
}

type loggingRow struct {
	pgx.Row
	ctx    context.Context
	start  time.Time
	method string
}

func (r *loggingRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err == pgx.ErrNoRows {
		err = nil
	}
	logger.LogExternalCall(r.ctx, "postgres", r.method, r.start, err)
	return err
}

func (p *LoggingPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	start := time.Now()
	row := p.next.QueryRow(ctx, sql, args...)
	return &loggingRow{Row: row, ctx: ctx, start: start, method: "query_row"}
}

func (p *LoggingPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	rows, err := p.next.Query(ctx, sql, args...)
	logger.LogExternalCall(ctx, "postgres", "query", start, err)
	return rows, err
}

func (p *LoggingPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	tag, err := p.next.Exec(ctx, sql, args...)
	logger.LogExternalCall(ctx, "postgres", "exec", start, err)
	return tag, err
}

func (p *LoggingPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.next.Begin(ctx)
}
