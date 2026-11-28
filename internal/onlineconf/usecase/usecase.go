package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
)

type Usecases struct {
	Scopes    ScopeUsecase
	Configs   ConfigUsecase
	Flags     FeatureFlagUsecase
	Overrides OverrideUsecase
	Keys      KeyUsecase
	Admins    AdminUsecase
	History   HistoryUsecase
	Snapshots SnapshotUsecase
}

func New(store *repository.Store) Usecases {
	return Usecases{
		Scopes: ScopeUsecase{store}, Configs: ConfigUsecase{store}, Flags: FeatureFlagUsecase{store},
		Overrides: OverrideUsecase{store}, Keys: KeyUsecase{store}, Admins: AdminUsecase{store},
		History: HistoryUsecase{store}, Snapshots: SnapshotUsecase{store},
	}
}

func checkAccess(admin models.Admin, err error, ownerOnly bool) (models.Admin, error) {
	if errors.Is(err, apperrors.ErrNotFound) {
		return models.Admin{}, apperrors.ErrForbidden
	}
	if err != nil {
		return models.Admin{}, err
	}
	if admin.Role == models.AdminRoleOwner || (!ownerOnly && admin.Role == models.AdminRoleAdmin) {
		return admin, nil
	}
	return models.Admin{}, apperrors.ErrForbidden
}

func protected[T any](ctx context.Context, store *repository.Store, actorID int64, ownerOnly bool,
	fn func(*repository.Store) (T, error),
) (T, error) {
	var result T
	err := store.WithinTx(ctx, func(tx *repository.Store) error {
		admin, err := tx.Admins.GetForShare(ctx, actorID)
		if _, err := checkAccess(admin, err, ownerOnly); err != nil {
			return err
		}
		result, err = fn(tx)
		return err
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

func protectedAction(ctx context.Context, store *repository.Store, actorID int64, ownerOnly bool,
	fn func(*repository.Store) error,
) error {
	_, err := protected(ctx, store, actorID, ownerOnly, func(tx *repository.Store) (struct{}, error) {
		return struct{}{}, fn(tx)
	})
	return err
}

type resource struct {
	kind  models.ResourceType
	id    int64
	scope *string
	name  string
}

func scopeResource(scope models.Scope) resource {
	return resource{kind: models.ResourceTypeScope, id: scope.ID, name: scope.Name, scope: &scope.Name}
}

func keyResource(kind models.ResourceType, key models.ConfigKey) resource {
	return resource{kind: kind, id: key.ID, name: key.Name, scope: &key.ScopeName}
}

func adminResource(userID int64) resource {
	return resource{kind: models.ResourceTypeAdmin, id: userID, name: strconv.FormatInt(userID, 10)}
}

func recordChange(ctx context.Context, tx *repository.Store, actorID int64, ref resource, before, after any) error {
	change := models.ChangeHistory{
		ActorUserID: actorID, ResourceType: ref.kind, ResourceID: ref.id, ScopeName: ref.scope, ResourceName: ref.name,
		Action: models.ActionUpdate,
	}
	var err error
	if before == nil {
		change.Action = models.ActionCreate
	} else {
		change.OldValue, err = json.Marshal(before)
		if err != nil {
			return err
		}
	}
	if after == nil {
		change.Action = models.ActionDelete
	} else {
		change.NewValue, err = json.Marshal(after)
		if err != nil {
			return err
		}
	}
	_, err = tx.History.Append(ctx, change)
	return err
}

func sameOptional[T comparable](a, b *T) bool {
	return a == b || (a != nil && b != nil && *a == *b)
}

func sameKey(a, b models.ConfigKey) bool {
	return a.ScopeID == b.ScopeID && a.Name == b.Name && a.OwnerUserID == b.OwnerUserID &&
		sameOptional(a.Description, b.Description)
}
