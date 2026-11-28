package usecase

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

type FeatureFlagUsecase struct {
	store Store
	users ProductUsers
}

func (u FeatureFlagUsecase) Get(ctx context.Context, actorID, keyID int64) (models.FeatureFlag, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, access keyAccess) (models.FeatureFlag, error) {
		if err := access.requireID(ctx, tx, models.ReadKey, keyID); err != nil {
			return models.FeatureFlag{}, err
		}
		return tx.Flags.Get(ctx, keyID)
	})
}

func (u FeatureFlagUsecase) Create(ctx context.Context, actorID int64, input models.FeatureFlag) (models.FeatureFlag, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, access keyAccess) (models.FeatureFlag, error) {
		if err := access.require(models.CreateKey, models.ConfigKey{ScopeID: input.ScopeID, Type: models.KeyTypeFeatureFlag}); err != nil {
			return models.FeatureFlag{}, err
		}
		if input.RolloutPercentage != 0 {
			return models.FeatureFlag{}, apperrors.ValidationError("new flags start at zero percent")
		}
		if err := requireUser(ctx, u.users, input.OwnerUserID); err != nil {
			return models.FeatureFlag{}, err
		}
		flag, err := tx.Flags.Create(ctx, input)
		if err != nil {
			return models.FeatureFlag{}, err
		}
		return flag, recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeFeatureFlag, flag.ConfigKey), nil, flag)
	})
}

func (u FeatureFlagUsecase) Update(ctx context.Context, actorID int64, input models.FeatureFlag) (models.FeatureFlag, error) {
	return u.update(ctx, actorID, input.ID, func(models.FeatureFlag) models.FeatureFlag { return input })
}

func (u FeatureFlagUsecase) UpdateMetadata(ctx context.Context, actorID int64, input models.FeatureFlag) (models.FeatureFlag, error) {
	return u.update(ctx, actorID, input.ID, func(before models.FeatureFlag) models.FeatureFlag {
		input.RolloutPercentage = before.RolloutPercentage
		return input
	})
}

func (u FeatureFlagUsecase) SetEnabled(ctx context.Context, actorID, keyID int64, enabled bool) (models.FeatureFlag, error) {
	return u.update(ctx, actorID, keyID, func(flag models.FeatureFlag) models.FeatureFlag {
		flag.Enabled = enabled
		return flag
	})
}

func (u FeatureFlagUsecase) SetRolloutPercentage(ctx context.Context, actorID, keyID int64, percentage int16) (models.FeatureFlag, error) {
	return u.update(ctx, actorID, keyID, func(flag models.FeatureFlag) models.FeatureFlag {
		flag.RolloutPercentage = percentage
		return flag
	})
}

func (u FeatureFlagUsecase) update(ctx context.Context, actorID, keyID int64,
	apply func(models.FeatureFlag) models.FeatureFlag,
) (models.FeatureFlag, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, access keyAccess) (models.FeatureFlag, error) {
		if err := access.requireID(ctx, tx, models.UpdateKey, keyID); err != nil {
			return models.FeatureFlag{}, err
		}
		before, err := tx.Flags.GetForUpdate(ctx, keyID)
		if err != nil {
			return models.FeatureFlag{}, err
		}
		input := apply(before)
		if err := access.requireMove(before.ConfigKey, input.ScopeID); err != nil {
			return models.FeatureFlag{}, err
		}
		if input.RolloutPercentage < 0 || input.RolloutPercentage > before.RolloutPercentage {
			return models.FeatureFlag{}, apperrors.ValidationError("increasing rollout requires an approved rollout plan")
		}
		if input.Enabled != before.Enabled {
			input.RolloutPercentage = 0
		}
		if input.OwnerUserID != before.OwnerUserID {
			if err := requireUser(ctx, u.users, input.OwnerUserID); err != nil {
				return models.FeatureFlag{}, err
			}
		}
		if sameKey(before.ConfigKey, input.ConfigKey) && (input.Type == "" || input.Type == before.Type) &&
			before.Enabled == input.Enabled && before.RolloutPercentage == input.RolloutPercentage {
			return before, nil
		}
		flag, err := tx.Flags.Update(ctx, input)
		if err != nil {
			return models.FeatureFlag{}, err
		}
		if input.Enabled != before.Enabled || input.RolloutPercentage != before.RolloutPercentage || input.ScopeID != before.ScopeID {
			if err := tx.Rollouts.CancelActive(ctx, keyID); err != nil {
				return models.FeatureFlag{}, err
			}
		}
		return flag, recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeFeatureFlag, flag.ConfigKey), before, flag)
	})
}

func (u FeatureFlagUsecase) Delete(ctx context.Context, actorID, keyID int64) error {
	return withAccessAction(ctx, u.store, actorID, func(tx Repositories, access keyAccess) error {
		if err := access.requireID(ctx, tx, models.DeleteKey, keyID); err != nil {
			return err
		}
		flag, err := tx.Flags.GetForUpdate(ctx, keyID)
		if err != nil {
			return err
		}
		overrides, err := tx.Overrides.ListAllByFlag(ctx, keyID)
		if err != nil {
			return err
		}
		if err := tx.Flags.Delete(ctx, keyID); err != nil {
			return err
		}
		before := flagDeletion{flag, overrides}
		return recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeFeatureFlag, flag.ConfigKey), before, nil)
	})
}

func (u FeatureFlagUsecase) Evaluate(ctx context.Context, actorID, keyID, userID int64) (onlineconf.Evaluation, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, access keyAccess) (onlineconf.Evaluation, error) {
		if err := access.requireID(ctx, tx, models.ReadKey, keyID); err != nil {
			return onlineconf.Evaluation{}, err
		}
		if userID <= 0 {
			return onlineconf.Evaluation{}, apperrors.ValidationError("user id must be positive")
		}
		flag, err := tx.Flags.Get(ctx, keyID)
		if err != nil {
			return onlineconf.Evaluation{}, err
		}
		if err := requireUser(ctx, u.users, userID); err != nil {
			return onlineconf.Evaluation{}, err
		}
		var enabledOverride *bool
		if flag.Enabled {
			override, err := tx.Overrides.Get(ctx, keyID, userID)
			switch {
			case err == nil:
				enabledOverride = &override.Enabled
			case !errors.Is(err, apperrors.ErrNotFound):
				return onlineconf.Evaluation{}, err
			}
		}
		return onlineconf.Evaluate(flag.ID, userID, flag.Enabled, flag.RolloutPercentage, enabledOverride), nil
	})
}
