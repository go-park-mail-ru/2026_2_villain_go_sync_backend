package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type AdminUsecase struct {
	store Store
	users ProductUsers
}

func (u AdminUsecase) CheckAccess(ctx context.Context, actorID int64) (models.Admin, error) {
	return withAccess(ctx, u.store, actorID, func(_ Repositories, access keyAccess) (models.Admin, error) {
		return access.admin, nil
	})
}

func (u AdminUsecase) List(ctx context.Context, actorID int64, page models.PageRequest) (models.Page[models.Admin], error) {
	return protected(ctx, u.store, actorID, ownerAccess, func(tx Repositories) (models.Page[models.Admin], error) {
		return tx.Admins.List(ctx, page)
	})
}

func (u AdminUsecase) Grant(ctx context.Context, actorID, userID int64, role models.AdminRole) (models.Admin, error) {
	return protected(ctx, u.store, actorID, changeAdmins, func(tx Repositories) (models.Admin, error) {
		if err := requireUser(ctx, u.users, userID); err != nil {
			return models.Admin{}, err
		}
		admin, err := tx.Admins.Create(ctx, models.Admin{UserID: userID, Role: role, GrantedBy: &actorID})
		if err != nil {
			return models.Admin{}, err
		}
		return admin, recordChange(ctx, tx, actorID, adminResource(admin.UserID), nil, admin)
	})
}

func (u AdminUsecase) ChangeRole(ctx context.Context, actorID, userID int64, role models.AdminRole) (models.Admin, error) {
	return protected(ctx, u.store, actorID, changeAdmins, func(tx Repositories) (models.Admin, error) {
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
	return protectedAction(ctx, u.store, actorID, changeAdmins, func(tx Repositories) error {
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
