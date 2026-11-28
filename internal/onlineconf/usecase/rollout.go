package usecase

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

type RolloutUsecase struct {
	store Store
	now   func() time.Time
}

func (u RolloutUsecase) ListPolicies(ctx context.Context, actorID int64) ([]models.RolloutPolicy, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, _ keyAccess) ([]models.RolloutPolicy, error) {
		return tx.Rollouts.ListPolicies(ctx)
	})
}

func (u RolloutUsecase) SavePolicy(ctx context.Context, actorID int64, policy models.RolloutPolicy) (models.RolloutPolicy, error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.RolloutPolicy, error) {
		policy.Name = strings.TrimSpace(policy.Name)
		if err := validatePolicy(policy); err != nil {
			return models.RolloutPolicy{}, err
		}
		return tx.Rollouts.SavePolicy(ctx, policy)
	})
}

func validatePolicy(policy models.RolloutPolicy) error {
	if policy.Name == "" || utf8.RuneCountInString(policy.Name) > 100 {
		return apperrors.ValidationError("a policy name is required")
	}
	if len(policy.Steps) < 2 || policy.Steps[len(policy.Steps)-1].Percentage != onlineconf.MaxRolloutPercentage {
		return apperrors.ValidationError("a policy must have at least two steps and end at 100 percent")
	}
	var previous int16
	for i, step := range policy.Steps {
		if step.Percentage <= previous || step.WaitSeconds < 0 || (i > 0 && step.WaitSeconds == 0) {
			return apperrors.ValidationError("percentages must increase and later steps require a positive wait")
		}
		previous = step.Percentage
	}
	return nil
}

func (u RolloutUsecase) DeletePolicy(ctx context.Context, actorID, policyID int64) error {
	return protectedAction(ctx, u.store, actorID, adminAccess, func(tx Repositories) error {
		return tx.Rollouts.DeletePolicy(ctx, policyID)
	})
}

func (u RolloutUsecase) Get(ctx context.Context, actorID, keyID int64) (*models.Rollout, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, access keyAccess) (*models.Rollout, error) {
		if err := access.requireID(ctx, tx, models.ReadKey, keyID); err != nil {
			return nil, err
		}
		if _, err := tx.Flags.Get(ctx, keyID); err != nil {
			return nil, err
		}
		rollout, err := tx.Rollouts.Get(ctx, keyID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, nil
		}
		return &rollout, err
	})
}

func (u RolloutUsecase) Start(ctx context.Context, actorID, keyID, policyID int64) (models.Rollout, error) {
	return u.change(ctx, actorID, keyID, models.UpdateKey, func(tx Repositories, flag models.FeatureFlag) (models.Rollout, error) {
		if !flag.Enabled || flag.RolloutPercentage >= onlineconf.MaxRolloutPercentage {
			return models.Rollout{}, apperrors.ErrConflict
		}
		policy, err := tx.Rollouts.GetPolicy(ctx, policyID)
		if err != nil {
			return models.Rollout{}, err
		}
		for index, step := range policy.Steps {
			if step.Percentage <= flag.RolloutPercentage {
				continue
			}
			dueAt := u.now().Add(time.Duration(step.WaitSeconds) * time.Second)
			rollout, err := tx.Rollouts.Start(ctx, models.Rollout{
				KeyID: keyID, PolicyName: policy.Name, Steps: policy.Steps, NextStep: index, DueAt: &dueAt,
			})
			if errors.Is(err, apperrors.ErrNotFound) {
				return models.Rollout{}, apperrors.ErrConflict
			}
			return rollout, err
		}
		return models.Rollout{}, apperrors.ErrConflict
	})
}

func (u RolloutUsecase) Approve(ctx context.Context, actorID, keyID, runID int64, step int) (models.Rollout, error) {
	return u.change(ctx, actorID, keyID, models.ApproveRollout, func(tx Repositories, _ models.FeatureFlag) (models.Rollout, error) {
		rollout, err := tx.Rollouts.Get(ctx, keyID)
		if err != nil {
			return models.Rollout{}, err
		}
		if rollout.Status != models.RolloutRunning || rollout.RunID != runID || rollout.NextStep != step {
			return models.Rollout{}, apperrors.ErrConflict
		}
		if rollout.ApprovedBy != nil {
			return rollout, nil
		}
		rollout.ApprovedBy = &actorID
		return tx.Rollouts.Save(ctx, rollout)
	})
}

func (u RolloutUsecase) Cancel(ctx context.Context, actorID, keyID, runID int64) (models.Rollout, error) {
	return u.change(ctx, actorID, keyID, models.UpdateKey, func(tx Repositories, _ models.FeatureFlag) (models.Rollout, error) {
		rollout, err := tx.Rollouts.Get(ctx, keyID)
		if err != nil {
			return models.Rollout{}, err
		}
		if rollout.RunID != runID || rollout.Status == models.RolloutCompleted {
			return models.Rollout{}, apperrors.ErrConflict
		}
		rollout.Status = models.RolloutCanceled
		rollout.DueAt, rollout.ApprovedBy = nil, nil
		return tx.Rollouts.Save(ctx, rollout)
	})
}

func (u RolloutUsecase) change(ctx context.Context, actorID, keyID int64, permission models.Permission,
	fn func(Repositories, models.FeatureFlag) (models.Rollout, error),
) (models.Rollout, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, access keyAccess) (models.Rollout, error) {
		if err := access.requireID(ctx, tx, permission, keyID); err != nil {
			return models.Rollout{}, err
		}
		flag, err := tx.Flags.GetForUpdate(ctx, keyID)
		if err != nil {
			return models.Rollout{}, err
		}
		return fn(tx, flag)
	})
}

func (u RolloutUsecase) ProcessReady(ctx context.Context) error {
	ids, err := u.store.Repositories().Rollouts.Ready(ctx, u.now())
	if err != nil {
		return err
	}
	var failed error
	for _, id := range ids {
		err := u.advance(ctx, id)
		if err != nil && !errors.Is(err, apperrors.ErrRetryable) && !errors.Is(err, apperrors.ErrNotFound) {
			failed = errors.Join(failed, err)
		}
	}
	return failed
}

func (u RolloutUsecase) advance(ctx context.Context, keyID int64) error {
	return u.store.WithinTransaction(ctx, func(tx Repositories) error {
		flag, err := tx.Flags.GetForUpdate(ctx, keyID)
		if err != nil {
			return err
		}
		rollout, err := tx.Rollouts.Get(ctx, keyID)
		if err != nil {
			return err
		}
		now := u.now()
		if rollout.Status != models.RolloutRunning || rollout.ApprovedBy == nil || rollout.DueAt.After(now) {
			return nil
		}
		if !flag.Enabled {
			return tx.Rollouts.CancelActive(ctx, keyID)
		}
		actorID := *rollout.ApprovedBy
		access, err := accessFor(ctx, tx, actorID)
		if err == nil {
			err = access.require(models.ApproveRollout, flag.ConfigKey)
		}
		rollout.ApprovedBy = nil
		if errors.Is(err, apperrors.ErrForbidden) {
			_, err = tx.Rollouts.Save(ctx, rollout)
			return err
		}
		if err != nil {
			return err
		}
		updated := flag
		updated.RolloutPercentage = rollout.Steps[rollout.NextStep].Percentage
		updated, err = tx.Flags.Update(ctx, updated)
		if err != nil {
			return err
		}
		if err := recordChange(ctx, tx, actorID, keyResource(models.ResourceTypeFeatureFlag, updated.ConfigKey), flag, updated); err != nil {
			return err
		}
		rollout.NextStep++
		if rollout.NextStep == len(rollout.Steps) {
			rollout.Status, rollout.DueAt = models.RolloutCompleted, nil
		} else {
			next := now.Add(time.Duration(rollout.Steps[rollout.NextStep].WaitSeconds) * time.Second)
			rollout.DueAt = &next
		}
		_, err = tx.Rollouts.Save(ctx, rollout)
		return err
	})
}
