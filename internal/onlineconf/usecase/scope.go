package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type ScopeUsecase struct {
	store Store
}

func (u ScopeUsecase) Get(ctx context.Context, actorID, scopeID int64) (models.Scope, error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, _ keyAccess) (models.Scope, error) {
		return tx.Scopes.Get(ctx, scopeID)
	})
}

func (u ScopeUsecase) List(ctx context.Context, actorID int64, page models.PageRequest) (models.Page[models.Scope], error) {
	return withAccess(ctx, u.store, actorID, func(tx Repositories, _ keyAccess) (models.Page[models.Scope], error) {
		return tx.Scopes.List(ctx, page)
	})
}

func (u ScopeUsecase) Create(ctx context.Context, actorID int64, input models.Scope) (models.Scope, error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.Scope, error) {
		scope, err := tx.Scopes.Create(ctx, input)
		if err != nil {
			return models.Scope{}, err
		}
		return scope, recordChange(ctx, tx, actorID, scopeResource(scope), nil, scope)
	})
}

func (u ScopeUsecase) Update(ctx context.Context, actorID int64, input models.Scope) (models.Scope, error) {
	return protected(ctx, u.store, actorID, adminAccess, func(tx Repositories) (models.Scope, error) {
		before, err := tx.Scopes.Get(ctx, input.ID)
		if err != nil {
			return models.Scope{}, err
		}
		if before.Name == input.Name && sameOptional(before.Description, input.Description) {
			return before, nil
		}
		scope, err := tx.Scopes.Update(ctx, input)
		if err != nil {
			return models.Scope{}, err
		}
		return scope, recordChange(ctx, tx, actorID, scopeResource(scope), before, scope)
	})
}

func (u ScopeUsecase) Delete(ctx context.Context, actorID, scopeID int64) error {
	return protectedAction(ctx, u.store, actorID, adminAccess, func(tx Repositories) error {
		before, err := tx.Scopes.Get(ctx, scopeID)
		if err != nil {
			return err
		}
		if err := tx.Scopes.Delete(ctx, scopeID); err != nil {
			return err
		}
		return recordChange(ctx, tx, actorID, scopeResource(before), before, nil)
	})
}
