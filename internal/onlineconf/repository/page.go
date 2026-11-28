package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/jackc/pgx/v5"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

type PageRequest struct {
	Cursor *int64
	Limit  int
}

type Page[T any] struct {
	Items      []T
	NextCursor *int64
}

func readPage[T any](ctx context.Context, db database, request PageRequest, query string,
	scan func(rowScanner) (T, error), id func(T) int64, filters ...any,
) (Page[T], error) {
	limit := request.Limit
	if limit == 0 {
		limit = defaultPageSize
	}
	if limit < 1 || limit > maxPageSize || (request.Cursor != nil && *request.Cursor <= 0) {
		return Page[T]{}, apperrors.ValidationError("invalid page limit or cursor")
	}

	args := append([]any{request.Cursor, limit + 1}, filters...)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return Page[T]{}, classifyError(ctx, err)
	}
	items, err := pgx.AppendRows(make([]T, 0, limit+1), rows, func(row pgx.CollectableRow) (T, error) {
		return scan(row)
	})
	if err != nil {
		return Page[T]{}, classifyError(ctx, err)
	}
	page := Page[T]{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		cursor := id(items[limit-1])
		page.NextCursor = &cursor
	}
	return page, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}
