package usecase

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type OverrideUsecase struct {
	store Store
	users ProductUsers
}

func (u OverrideUsecase) Get(ctx context.Context, actorID, keyID, userID int64) (models.FeatureFlagOverride, error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.FeatureFlagOverride, error) {
		return tx.Overrides.Get(ctx, keyID, userID)
	})
}

func (u OverrideUsecase) ListByFlag(ctx context.Context, actorID, keyID int64,
	page models.PageRequest,
) (models.Page[models.FeatureFlagOverride], error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.Page[models.FeatureFlagOverride], error) {
		if _, err := tx.Flags.Get(ctx, keyID); err != nil {
			return models.Page[models.FeatureFlagOverride]{}, err
		}
		return tx.Overrides.ListByFlag(ctx, keyID, page)
	})
}

func (u OverrideUsecase) ListByUser(ctx context.Context, actorID, userID int64,
	page models.PageRequest,
) (models.Page[models.FeatureFlagOverride], error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.Page[models.FeatureFlagOverride], error) {
		return tx.Overrides.ListByUser(ctx, userID, page)
	})
}

func (u OverrideUsecase) Upsert(ctx context.Context, actorID int64, input models.FeatureFlagOverride) (models.FeatureFlagOverride, error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.FeatureFlagOverride, error) {
		flag, err := tx.Flags.Get(ctx, input.KeyID)
		if err != nil {
			return models.FeatureFlagOverride{}, err
		}
		var before any
		old, err := tx.Overrides.Get(ctx, input.KeyID, input.UserID)
		switch {
		case err == nil:
			if old.Enabled == input.Enabled {
				return old, nil
			}
			before = old
		case !errors.Is(err, apperrors.ErrNotFound):
			return models.FeatureFlagOverride{}, err
		}
		if before == nil {
			if err := requireUser(ctx, u.users, input.UserID); err != nil {
				return models.FeatureFlagOverride{}, err
			}
		}
		// Lock the parent before the trigger can acquire it from a child update.
		flag, err = tx.Flags.GetForUpdate(ctx, input.KeyID)
		if err != nil {
			return models.FeatureFlagOverride{}, err
		}
		override, err := tx.Overrides.Upsert(ctx, input)
		if err != nil {
			return models.FeatureFlagOverride{}, err
		}
		return override, recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeFeatureFlagOverride, flag.ConfigKey), before, override)
	})
}

func (u OverrideUsecase) Delete(ctx context.Context, actorID, keyID, userID int64) error {
	return protectedAction(ctx, u.store, actorID, adminAccess, func(tx Repositories) error {
		flag, err := tx.Flags.GetForUpdate(ctx, keyID)
		if err != nil {
			return err
		}
		before, err := tx.Overrides.Get(ctx, keyID, userID)
		if err != nil {
			return err
		}
		if err := tx.Overrides.Delete(ctx, keyID, userID); err != nil {
			return err
		}
		return recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeFeatureFlagOverride, flag.ConfigKey), before, nil)
	})
}
