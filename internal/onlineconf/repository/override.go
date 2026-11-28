package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/jackc/pgx/v5"
)

type FeatureFlagOverrideRepository struct {
	db database
}

const overrideColumns = "key_id, user_id, enabled, created_at, updated_at"

func scanOverride(row rowScanner) (models.FeatureFlagOverride, error) {
	var override models.FeatureFlagOverride
	err := row.Scan(&override.KeyID, &override.UserID, &override.Enabled, &override.CreatedAt, &override.UpdatedAt)
	return override, err
}

func (r FeatureFlagOverrideRepository) Get(ctx context.Context, keyID, userID int64) (models.FeatureFlagOverride, error) {
	override, err := scanOverride(r.db.QueryRow(ctx, `SELECT `+overrideColumns+` FROM public.feature_flag_override
		WHERE key_id = $1 AND user_id = $2`, keyID, userID))
	return override, classifyError(ctx, err)
}

func (r FeatureFlagOverrideRepository) Upsert(ctx context.Context, input models.FeatureFlagOverride) (models.FeatureFlagOverride, error) {
	override, err := scanOverride(r.db.QueryRow(ctx, `
		INSERT INTO public.feature_flag_override (key_id, user_id, enabled) VALUES ($1, $2, $3)
		ON CONFLICT (key_id, user_id) DO UPDATE SET enabled = EXCLUDED.enabled
		RETURNING `+overrideColumns, input.KeyID, input.UserID, input.Enabled))
	return override, classifyError(ctx, err)
}

func (r FeatureFlagOverrideRepository) Delete(ctx context.Context, keyID, userID int64) error {
	return execOne(ctx, r.db, "DELETE FROM public.feature_flag_override WHERE key_id = $1 AND user_id = $2", keyID, userID)
}

func (r FeatureFlagOverrideRepository) ListByFlag(ctx context.Context, keyID int64, page models.PageRequest) (models.Page[models.FeatureFlagOverride], error) {
	return readPage(ctx, r.db, page, `SELECT `+overrideColumns+` FROM public.feature_flag_override
		WHERE key_id = $3 AND ($1::BIGINT IS NULL OR user_id > $1) ORDER BY user_id LIMIT $2`, scanOverride,
		func(override models.FeatureFlagOverride) int64 { return override.UserID }, keyID)
}

func (r FeatureFlagOverrideRepository) ListByUser(ctx context.Context, userID int64, page models.PageRequest) (models.Page[models.FeatureFlagOverride], error) {
	return readPage(ctx, r.db, page, `SELECT `+overrideColumns+` FROM public.feature_flag_override
		WHERE user_id = $3 AND ($1::BIGINT IS NULL OR key_id > $1) ORDER BY key_id LIMIT $2`, scanOverride,
		func(override models.FeatureFlagOverride) int64 { return override.KeyID }, userID)
}

func (r FeatureFlagOverrideRepository) ListAllByFlag(ctx context.Context, keyID int64) ([]models.FeatureFlagOverride, error) {
	rows, err := r.db.Query(ctx, `SELECT `+overrideColumns+` FROM public.feature_flag_override
		WHERE key_id = $1 ORDER BY user_id`, keyID)
	if err != nil {
		return nil, classifyError(ctx, err)
	}
	items, err := pgx.AppendRows([]models.FeatureFlagOverride{}, rows, func(row pgx.CollectableRow) (models.FeatureFlagOverride, error) {
		return scanOverride(row)
	})
	return items, classifyError(ctx, err)
}
