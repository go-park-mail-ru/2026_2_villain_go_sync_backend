package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type KeyUsecase struct {
	store Store
}

func (u KeyUsecase) Get(ctx context.Context, actorID, keyID int64) (models.KeyDetails, error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.KeyDetails, error) {
		return tx.Keys.GetDetails(ctx, keyID)
	})
}

func (u KeyUsecase) List(ctx context.Context, actorID int64, filter models.KeyFilter) (models.Page[models.KeyDetails], error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.Page[models.KeyDetails], error) {
		return tx.Keys.List(ctx, filter)
	})
}

type HistoryUsecase struct {
	store Store
}

func (u HistoryUsecase) Get(ctx context.Context, actorID, changeID int64) (models.ChangeHistory, error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.ChangeHistory, error) {
		return tx.History.Get(ctx, changeID)
	})
}

func (u HistoryUsecase) List(ctx context.Context, actorID int64, filter models.HistoryFilter) (models.Page[models.ChangeHistory], error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.Page[models.ChangeHistory], error) {
		return tx.History.List(ctx, filter)
	})
}

type SnapshotUsecase struct {
	store Store
}

func (u SnapshotUsecase) Get(ctx context.Context) (models.Snapshot, error) {
	return u.store.Repositories().Snapshots.Get(ctx)
}
