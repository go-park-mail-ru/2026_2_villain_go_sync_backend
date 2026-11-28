package usecase

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type keyAccess struct {
	models.Access
	admin models.Admin
}

func accessFor(ctx context.Context, tx Repositories, actorID int64) (keyAccess, error) {
	admin, err := tx.Admins.GetForShare(ctx, actorID)
	if !errors.Is(err, apperrors.ErrNotFound) {
		admin, err = checkAccess(admin, err, false)
		return keyAccess{Access: models.Access{Full: true, Grants: []models.Grant{}}, admin: admin}, err
	}
	grants, err := tx.Access.Effective(ctx, actorID)
	if err != nil {
		return keyAccess{}, err
	}
	if len(grants) == 0 {
		return keyAccess{}, apperrors.ErrForbidden
	}
	return keyAccess{
		Access: models.Access{Grants: grants},
		admin:  models.Admin{UserID: actorID, Role: models.AdminRoleMember},
	}, nil
}

func withAccess[T any](ctx context.Context, store Store, actorID int64, fn func(Repositories, keyAccess) (T, error)) (T, error) {
	var result T
	err := store.WithinTransaction(ctx, func(tx Repositories) error {
		access, err := accessFor(ctx, tx, actorID)
		if err != nil {
			return err
		}
		result, err = fn(tx, access)
		return err
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

func withAccessAction(ctx context.Context, store Store, actorID int64, fn func(Repositories, keyAccess) error) error {
	_, err := withAccess(ctx, store, actorID, func(tx Repositories, access keyAccess) (struct{}, error) {
		return struct{}{}, fn(tx, access)
	})
	return err
}

func (a keyAccess) require(action models.Permission, key models.ConfigKey) error {
	if a.Full {
		return nil
	}
	for _, grant := range a.Grants {
		if grant.Action != action {
			continue
		}
		if grant.KeyID != nil && *grant.KeyID == key.ID {
			return nil
		}
		if grant.ScopeID != nil && *grant.ScopeID == key.ScopeID && *grant.KeyType == key.Type {
			return nil
		}
	}
	return apperrors.ErrForbidden
}

func (a keyAccess) requireID(ctx context.Context, tx Repositories, action models.Permission, id int64) error {
	if a.Full {
		return nil
	}
	key, err := tx.Keys.Get(ctx, id)
	if errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.ErrForbidden
	}
	if err != nil {
		return err
	}
	return a.require(action, key)
}

func (a keyAccess) requireMove(before models.ConfigKey, scopeID int64) error {
	if scopeID == before.ScopeID {
		return nil
	}
	return a.require(models.CreateKey, models.ConfigKey{ScopeID: scopeID, Type: before.Type})
}

func (a keyAccess) readable() *models.ReadableKeys {
	if a.Full {
		return nil
	}
	keys := &models.ReadableKeys{}
	for _, grant := range a.Grants {
		if grant.Action != models.ReadKey {
			continue
		}
		switch {
		case grant.KeyID != nil:
			keys.IDs = append(keys.IDs, *grant.KeyID)
		case *grant.KeyType == models.KeyTypeConfig:
			keys.ConfigScopes = append(keys.ConfigScopes, *grant.ScopeID)
		default:
			keys.FlagScopes = append(keys.FlagScopes, *grant.ScopeID)
		}
	}
	return keys
}

type AccessUsecase struct {
	store Store
	users ProductUsers
}

func (u AccessUsecase) Current(ctx context.Context, actorID int64) (models.Access, error) {
	return withAccess(ctx, u.store, actorID, func(_ Repositories, access keyAccess) (models.Access, error) {
		return access.Access, nil
	})
}

func (u AccessUsecase) ListGrants(ctx context.Context, actorID int64) ([]models.Grant, error) {
	return protected(ctx, u.store, actorID, ownerAccess, func(tx Repositories) ([]models.Grant, error) {
		return tx.Access.List(ctx)
	})
}

func (u AccessUsecase) Grant(ctx context.Context, actorID int64, grant models.Grant) (models.Grant, error) {
	return protected(ctx, u.store, actorID, ownerAccess, func(tx Repositories) (models.Grant, error) {
		if err := validateGrant(grant); err != nil {
			return models.Grant{}, err
		}
		if grant.UserID != nil {
			if err := requireUser(ctx, u.users, *grant.UserID); err != nil {
				return models.Grant{}, err
			}
		}
		if grant.KeyID != nil && grant.Action == models.ApproveRollout {
			key, err := tx.Keys.Get(ctx, *grant.KeyID)
			if err != nil {
				return models.Grant{}, err
			}
			if key.Type != models.KeyTypeFeatureFlag {
				return models.Grant{}, apperrors.ValidationError("only feature flags can be approved")
			}
		}
		return tx.Access.Create(ctx, grant)
	})
}

func validateGrant(grant models.Grant) error {
	switch grant.Action {
	case models.ReadKey, models.CreateKey, models.UpdateKey, models.DeleteKey, models.ApproveRollout:
	default:
		return apperrors.ValidationError("unknown permission")
	}
	if (grant.UserID == nil) == (grant.GroupID == nil) {
		return apperrors.ValidationError("specify one user or group")
	}
	for _, id := range []*int64{grant.UserID, grant.GroupID, grant.KeyID, grant.ScopeID} {
		if id != nil && *id <= 0 {
			return apperrors.ValidationError("identifiers must be positive")
		}
	}
	if grant.KeyID != nil {
		if grant.ScopeID != nil || grant.KeyType != nil || grant.Action == models.CreateKey {
			return apperrors.ValidationError("resource permissions must address an existing key")
		}
		return nil
	}
	if grant.ScopeID == nil || grant.KeyType == nil ||
		(*grant.KeyType != models.KeyTypeConfig && *grant.KeyType != models.KeyTypeFeatureFlag) {
		return apperrors.ValidationError("scope permissions require a scope and a key type")
	}
	if grant.Action == models.ApproveRollout && *grant.KeyType != models.KeyTypeFeatureFlag {
		return apperrors.ValidationError("only feature flags can be approved")
	}
	return nil
}

func (u AccessUsecase) Revoke(ctx context.Context, actorID, id int64) error {
	return protectedAction(ctx, u.store, actorID, ownerAccess, func(tx Repositories) error {
		return tx.Access.Delete(ctx, id)
	})
}

func (u AccessUsecase) ListGroups(ctx context.Context, actorID int64) ([]models.Group, error) {
	return protected(ctx, u.store, actorID, ownerAccess, func(tx Repositories) ([]models.Group, error) {
		return tx.Groups.List(ctx)
	})
}

func (u AccessUsecase) SaveGroup(ctx context.Context, actorID int64, group models.Group) (models.Group, error) {
	return protected(ctx, u.store, actorID, ownerAccess, func(tx Repositories) (models.Group, error) {
		group.Name = strings.TrimSpace(group.Name)
		if group.Name == "" || utf8.RuneCountInString(group.Name) > 100 || group.MemberIDs == nil {
			return models.Group{}, apperrors.ValidationError("group name and member list are required")
		}
		seen := make(map[int64]struct{}, len(group.MemberIDs))
		for _, userID := range group.MemberIDs {
			if _, duplicate := seen[userID]; duplicate {
				continue
			}
			if err := requireUser(ctx, u.users, userID); err != nil {
				return models.Group{}, err
			}
			seen[userID] = struct{}{}
		}
		return tx.Groups.Save(ctx, group)
	})
}

func (u AccessUsecase) DeleteGroup(ctx context.Context, actorID, id int64) error {
	return protectedAction(ctx, u.store, actorID, ownerAccess, func(tx Repositories) error {
		return tx.Groups.Delete(ctx, id)
	})
}
