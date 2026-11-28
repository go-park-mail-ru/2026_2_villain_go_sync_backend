package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
)

type ConfigUsecase struct {
	store *repository.Store
}

func (u ConfigUsecase) Get(ctx context.Context, actorID, keyID int64) (models.Config, error) {
	return protected(ctx, u.store, actorID, false, func(tx *repository.Store) (models.Config, error) {
		return tx.Configs.Get(ctx, keyID)
	})
}

func (u ConfigUsecase) Create(ctx context.Context, actorID int64, input models.Config) (models.Config, error) {
	return protected(ctx, u.store, actorID, false, func(tx *repository.Store) (models.Config, error) {
		config, err := tx.Configs.Create(ctx, input)
		if err != nil {
			return models.Config{}, err
		}
		return config, recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeConfig, config.ConfigKey), nil, config)
	})
}

func (u ConfigUsecase) Update(ctx context.Context, actorID int64, input models.Config) (models.Config, error) {
	return protected(ctx, u.store, actorID, false, func(tx *repository.Store) (models.Config, error) {
		before, err := tx.Configs.Get(ctx, input.ID)
		if err != nil {
			return models.Config{}, err
		}
		if sameKey(before.ConfigKey, input.ConfigKey) && (input.Type == "" || input.Type == before.Type) {
			equal, err := tx.Configs.ValueEquals(ctx, before.ID, input.Value)
			if err != nil {
				return models.Config{}, err
			}
			if equal {
				return before, nil
			}
		}
		config, err := tx.Configs.Update(ctx, input)
		if err != nil {
			return models.Config{}, err
		}
		return config, recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeConfig, config.ConfigKey), before, config)
	})
}

func (u ConfigUsecase) Delete(ctx context.Context, actorID, keyID int64) error {
	return protectedAction(ctx, u.store, actorID, false, func(tx *repository.Store) error {
		before, err := tx.Configs.Get(ctx, keyID)
		if err != nil {
			return err
		}
		if err := tx.Configs.Delete(ctx, keyID); err != nil {
			return err
		}
		return recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeConfig, before.ConfigKey), before, nil)
	})
}
