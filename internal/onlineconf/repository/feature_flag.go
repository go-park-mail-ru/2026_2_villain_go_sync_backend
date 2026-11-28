package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/jackc/pgx/v5"
)

type FeatureFlagRepository struct {
	conn connection
}

func (r FeatureFlagRepository) Get(ctx context.Context, keyID int64) (models.FeatureFlag, error) {
	return getFlag(ctx, r.conn.db, keyID)
}

func getFlag(ctx context.Context, db database, keyID int64) (models.FeatureFlag, error) {
	var flag models.FeatureFlag
	err := db.QueryRow(ctx, "SELECT "+keyColumns+", f.enabled, f.rollout_percentage"+keyFrom+
		" JOIN public.feature_flag f ON f.key_id = k.key_id WHERE k.key_id = $1", keyID).
		Scan(append(keyFields(&flag.ConfigKey), &flag.Enabled, &flag.RolloutPercentage)...)
	return flag, classifyError(ctx, err)
}

func (r FeatureFlagRepository) Create(ctx context.Context, input models.FeatureFlag) (models.FeatureFlag, error) {
	var flag models.FeatureFlag
	err := r.conn.withinTx(ctx, func(tx pgx.Tx) error {
		id, err := createKey(ctx, tx, input.ConfigKey, models.KeyTypeFeatureFlag)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO public.feature_flag (key_id, enabled, rollout_percentage) VALUES ($1, $2, $3)`,
			id, input.Enabled, input.RolloutPercentage); err != nil {
			return err
		}
		flag, err = getFlag(ctx, tx, id)
		return err
	})
	if err != nil {
		return models.FeatureFlag{}, err
	}
	return flag, nil
}

func (r FeatureFlagRepository) Update(ctx context.Context, input models.FeatureFlag) (models.FeatureFlag, error) {
	var flag models.FeatureFlag
	err := r.conn.withinTx(ctx, func(tx pgx.Tx) error {
		if err := updateKey(ctx, tx, input.ConfigKey, models.KeyTypeFeatureFlag); err != nil {
			return err
		}
		if err := execOne(ctx, tx, `
			UPDATE public.feature_flag SET enabled = $2, rollout_percentage = $3 WHERE key_id = $1`,
			input.ID, input.Enabled, input.RolloutPercentage); err != nil {
			return err
		}
		var err error
		flag, err = getFlag(ctx, tx, input.ID)
		return err
	})
	if err != nil {
		return models.FeatureFlag{}, err
	}
	return flag, nil
}

func (r FeatureFlagRepository) Delete(ctx context.Context, keyID int64) error {
	return deleteKey(ctx, r.conn.db, keyID, models.KeyTypeFeatureFlag)
}
