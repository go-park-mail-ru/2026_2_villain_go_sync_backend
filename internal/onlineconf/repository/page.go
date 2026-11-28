package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/jackc/pgx/v5"
)

func readPage[T any](ctx context.Context, db database, request models.PageRequest, query string,
	scan func(rowScanner) (T, error), id func(T) int64, filters ...any,
) (models.Page[T], error) {
	limit := request.Limit
	if limit == 0 {
		limit = models.DefaultPageSize
	}
	if limit < 1 || limit > models.MaxPageSize || (request.Cursor != nil && *request.Cursor <= 0) {
		return models.Page[T]{}, apperrors.ValidationError("invalid page limit or cursor")
	}

	args := append([]any{request.Cursor, limit + 1}, filters...)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return models.Page[T]{}, classifyError(ctx, err)
	}
	items, err := pgx.AppendRows(make([]T, 0, limit+1), rows, func(row pgx.CollectableRow) (T, error) {
		return scan(row)
	})
	if err != nil {
		return models.Page[T]{}, classifyError(ctx, err)
	}
	page := models.Page[T]{Items: items}
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
