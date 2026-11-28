package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
)

type KeyUsecase struct {
	store *repository.Store
}

func (u KeyUsecase) Get(ctx context.Context, actorID, keyID int64) (models.ConfigKey, error) {
	return protected(ctx, u.store, actorID, false, func(tx *repository.Store) (models.ConfigKey, error) {
		return tx.Keys.Get(ctx, keyID)
	})
}

func (u KeyUsecase) List(ctx context.Context, actorID int64, filter repository.KeyFilter) (repository.Page[models.ConfigKey], error) {
	return protected(ctx, u.store, actorID, false, func(tx *repository.Store) (repository.Page[models.ConfigKey], error) {
		return tx.Keys.List(ctx, filter)
	})
}

type HistoryUsecase struct {
	store *repository.Store
}

func (u HistoryUsecase) Get(ctx context.Context, actorID, changeID int64) (models.ChangeHistory, error) {
	return protected(ctx, u.store, actorID, false, func(tx *repository.Store) (models.ChangeHistory, error) {
		return tx.History.Get(ctx, changeID)
	})
}

func (u HistoryUsecase) List(ctx context.Context, actorID int64, filter repository.HistoryFilter) (repository.Page[models.ChangeHistory], error) {
	return protected(ctx, u.store, actorID, false, func(tx *repository.Store) (repository.Page[models.ChangeHistory], error) {
		return tx.History.List(ctx, filter)
	})
}

type SnapshotUsecase struct {
	store *repository.Store
}

func (u SnapshotUsecase) Get(ctx context.Context) (repository.Snapshot, error) {
	return u.store.Snapshots.Get(ctx)
}
