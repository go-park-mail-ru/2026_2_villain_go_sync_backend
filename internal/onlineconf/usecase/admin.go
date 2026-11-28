package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
)

type AdminUsecase struct {
	store *repository.Store
}

func (u AdminUsecase) CheckAccess(ctx context.Context, actorID int64) (models.Admin, error) {
	admin, err := u.store.Admins.Get(ctx, actorID)
	return checkAccess(admin, err, false)
}

func (u AdminUsecase) List(ctx context.Context, actorID int64, page repository.PageRequest) (repository.Page[models.Admin], error) {
	return protected(ctx, u.store, actorID, true, func(tx *repository.Store) (repository.Page[models.Admin], error) {
		return tx.Admins.List(ctx, page)
	})
}

func (u AdminUsecase) Grant(ctx context.Context, actorID, userID int64, role models.AdminRole) (models.Admin, error) {
	return protected(ctx, u.store, actorID, true, func(tx *repository.Store) (models.Admin, error) {
		admin, err := tx.Admins.Create(ctx, models.Admin{UserID: userID, Role: role, GrantedBy: &actorID})
		if err != nil {
			return models.Admin{}, err
		}
		return admin, recordChange(ctx, tx, actorID, adminResource(admin.UserID), nil, admin)
	})
}

func (u AdminUsecase) ChangeRole(ctx context.Context, actorID, userID int64, role models.AdminRole) (models.Admin, error) {
	return protected(ctx, u.store, actorID, true, func(tx *repository.Store) (models.Admin, error) {
		before, err := tx.Admins.Get(ctx, userID)
		if err != nil {
			return models.Admin{}, err
		}
		if before.Role == role {
			return before, nil
		}
		var grantedBy *int64
		if before.GrantedBy == nil && role == models.AdminRoleAdmin {
			grantedBy = &actorID
		}
		admin, err := tx.Admins.UpdateRole(ctx, userID, role, grantedBy)
		if err != nil {
			return models.Admin{}, err
		}
		return admin, recordChange(ctx, tx, actorID, adminResource(admin.UserID), before, admin)
	})
}

func (u AdminUsecase) Revoke(ctx context.Context, actorID, userID int64) error {
	return protectedAction(ctx, u.store, actorID, true, func(tx *repository.Store) error {
		before, err := tx.Admins.Get(ctx, userID)
		if err != nil {
			return err
		}
		if err := tx.Admins.Delete(ctx, userID); err != nil {
			return err
		}
		return recordChange(ctx, tx, actorID, adminResource(before.UserID), before, nil)
	})
}
