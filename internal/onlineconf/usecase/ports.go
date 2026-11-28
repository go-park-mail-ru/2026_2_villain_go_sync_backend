package usecase

import (
	"context"
	"encoding/json"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type Store interface {
	Repositories() Repositories
	WithinTransaction(context.Context, func(Repositories) error) error
}
type Repositories struct {
	Scopes    scopeRepository
	Configs   configRepository
	Flags     featureFlagRepository
	Overrides featureFlagOverrideRepository
	Admins    adminRepository
	History   historyRepository
	Keys      keyRepository
	Snapshots snapshotRepository
}
type scopeRepository interface {
	Create(ctx context.Context, scope models.Scope) (models.Scope, error)
	Get(ctx context.Context, scopeID int64) (models.Scope, error)
	List(ctx context.Context, page models.PageRequest) (models.Page[models.Scope], error)
	Update(ctx context.Context, scope models.Scope) (models.Scope, error)
	Delete(ctx context.Context, scopeID int64) error
}
type configRepository interface {
	Get(ctx context.Context, keyID int64) (models.Config, error)
	Create(ctx context.Context, input models.Config) (models.Config, error)
	Update(ctx context.Context, input models.Config) (models.Config, error)
	Delete(ctx context.Context, keyID int64) error
	ValueEquals(ctx context.Context, keyID int64, value json.RawMessage) (bool, error)
}
type featureFlagRepository interface {
	GetForUpdate(context.Context, int64) (models.FeatureFlag, error)
	Get(ctx context.Context, keyID int64) (models.FeatureFlag, error)
	Create(ctx context.Context, input models.FeatureFlag) (models.FeatureFlag, error)
	Update(ctx context.Context, input models.FeatureFlag) (models.FeatureFlag, error)
	Delete(ctx context.Context, keyID int64) error
}
type featureFlagOverrideRepository interface {
	Get(ctx context.Context, keyID, userID int64) (models.FeatureFlagOverride, error)
	Upsert(ctx context.Context, input models.FeatureFlagOverride) (models.FeatureFlagOverride, error)
	Delete(ctx context.Context, keyID, userID int64) error
	ListByFlag(ctx context.Context, keyID int64, page models.PageRequest) (models.Page[models.FeatureFlagOverride], error)
	ListByUser(ctx context.Context, userID int64, page models.PageRequest) (models.Page[models.FeatureFlagOverride], error)
	ListAllByFlag(ctx context.Context, keyID int64) ([]models.FeatureFlagOverride, error)
}
type adminRepository interface {
	LockChanges(context.Context) error
	Get(ctx context.Context, userID int64) (models.Admin, error)
	List(ctx context.Context, page models.PageRequest) (models.Page[models.Admin], error)
	Create(ctx context.Context, input models.Admin) (models.Admin, error)
	UpdateRole(ctx context.Context, userID int64, role models.AdminRole, grantedBy *int64) (models.Admin, error)
	Delete(ctx context.Context, userID int64) error
	GetForShare(ctx context.Context, userID int64) (models.Admin, error)
}
type historyRepository interface {
	Append(ctx context.Context, input models.ChangeHistory) (models.ChangeHistory, error)
	Get(ctx context.Context, changeID int64) (models.ChangeHistory, error)
	List(ctx context.Context, filter models.HistoryFilter) (models.Page[models.ChangeHistory], error)
}
type keyRepository interface {
	GetDetails(ctx context.Context, keyID int64) (models.KeyDetails, error)
	List(ctx context.Context, filter models.KeyFilter) (models.Page[models.KeyDetails], error)
}
type snapshotRepository interface {
	Get(ctx context.Context) (models.Snapshot, error)
}
