package repository

import (
	"context"
	"encoding/json"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/jackc/pgx/v5"
)

type ConfigRepository struct {
	conn connection
}

func (r ConfigRepository) Get(ctx context.Context, keyID int64) (models.Config, error) {
	return getConfig(ctx, r.conn.db, keyID)
}

func getConfig(ctx context.Context, db database, keyID int64) (models.Config, error) {
	var config models.Config
	err := db.QueryRow(ctx, "SELECT "+keyColumns+", c.value"+keyFrom+
		" JOIN public.config c ON c.key_id = k.key_id WHERE k.key_id = $1", keyID).
		Scan(append(keyFields(&config.ConfigKey), &config.Value)...)
	return config, classifyError(ctx, err)
}

func (r ConfigRepository) Create(ctx context.Context, input models.Config) (models.Config, error) {
	if !json.Valid(input.Value) {
		return models.Config{}, apperrors.ValidationError("config value must be valid JSON")
	}

	var config models.Config
	err := r.conn.withinTx(ctx, func(tx pgx.Tx) error {
		id, err := createKey(ctx, tx, input.ConfigKey, models.KeyTypeConfig)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "INSERT INTO public.config (key_id, value) VALUES ($1, $2)", id, input.Value); err != nil {
			return err
		}
		config, err = getConfig(ctx, tx, id)
		return err
	})
	if err != nil {
		return models.Config{}, err
	}
	return config, nil
}

func (r ConfigRepository) Update(ctx context.Context, input models.Config) (models.Config, error) {
	if !json.Valid(input.Value) {
		return models.Config{}, apperrors.ValidationError("config value must be valid JSON")
	}

	var config models.Config
	err := r.conn.withinTx(ctx, func(tx pgx.Tx) error {
		if err := updateKey(ctx, tx, input.ConfigKey, models.KeyTypeConfig); err != nil {
			return err
		}
		if err := execOne(ctx, tx, "UPDATE public.config SET value = $2 WHERE key_id = $1", input.ID, input.Value); err != nil {
			return err
		}
		var err error
		config, err = getConfig(ctx, tx, input.ID)
		return err
	})
	if err != nil {
		return models.Config{}, err
	}
	return config, nil
}

func (r ConfigRepository) Delete(ctx context.Context, keyID int64) error {
	return deleteKey(ctx, r.conn.db, keyID, models.KeyTypeConfig)
}

func (r ConfigRepository) ValueEquals(ctx context.Context, keyID int64, value json.RawMessage) (bool, error) {
	if !json.Valid(value) {
		return false, apperrors.ValidationError("config value must be valid JSON")
	}
	var equal bool
	err := r.conn.db.QueryRow(ctx, "SELECT value = $2::JSONB FROM public.config WHERE key_id = $1", keyID, value).Scan(&equal)
	return equal, classifyError(ctx, err)
}
