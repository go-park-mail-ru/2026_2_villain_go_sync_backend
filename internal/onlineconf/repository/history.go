package repository

import (
	"context"
	"encoding/json"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type HistoryRepository struct {
	db database
}

const historyColumns = "change_id, actor_user_id, resource_type, resource_id, scope_name, resource_name, action, old_value, new_value, created_at"

func scanHistory(row rowScanner) (models.ChangeHistory, error) {
	var change models.ChangeHistory
	err := row.Scan(&change.ID, &change.ActorUserID, &change.ResourceType, &change.ResourceID, &change.ScopeName,
		&change.ResourceName, &change.Action, &change.OldValue, &change.NewValue, &change.CreatedAt)
	return change, err
}

func (r HistoryRepository) Append(ctx context.Context, input models.ChangeHistory) (models.ChangeHistory, error) {
	if (input.OldValue != nil && !json.Valid(input.OldValue)) || (input.NewValue != nil && !json.Valid(input.NewValue)) {
		return models.ChangeHistory{}, apperrors.ValidationError("history values must be valid JSON or absent")
	}
	change, err := scanHistory(r.db.QueryRow(ctx, `
		INSERT INTO public.change_history
		    (actor_user_id, resource_type, resource_id, scope_name, resource_name, action, old_value, new_value)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING `+historyColumns,
		input.ActorUserID, input.ResourceType, input.ResourceID, input.ScopeName, input.ResourceName, input.Action,
		input.OldValue, input.NewValue))
	return change, classifyError(ctx, err)
}

func (r HistoryRepository) Get(ctx context.Context, changeID int64) (models.ChangeHistory, error) {
	change, err := scanHistory(r.db.QueryRow(ctx, `SELECT `+historyColumns+` FROM public.change_history WHERE change_id = $1`, changeID))
	return change, classifyError(ctx, err)
}

func (r HistoryRepository) List(ctx context.Context, filter models.HistoryFilter) (models.Page[models.ChangeHistory], error) {
	return readPage(ctx, r.db, filter.Page, `SELECT `+historyColumns+` FROM public.change_history
		WHERE ($1::BIGINT IS NULL OR change_id < $1)
		  AND ($3::TEXT IS NULL OR resource_type = $3)
		  AND ($4::BIGINT IS NULL OR resource_id = $4)
		  AND ($5::TEXT IS NULL OR scope_name = $5)
		ORDER BY change_id DESC LIMIT $2`, scanHistory,
		func(change models.ChangeHistory) int64 { return change.ID }, filter.ResourceType, filter.ResourceID, filter.ScopeName)
}
