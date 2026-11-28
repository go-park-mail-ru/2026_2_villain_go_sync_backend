package usecase

import (
	"context"
	"encoding/json"
	"time"

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
	Access    accessRepository
	Groups    groupRepository
	Rollouts  rolloutRepository
}

type accessRepository interface {
	Effective(context.Context, int64) ([]models.Grant, error)
	List(context.Context) ([]models.Grant, error)
	Create(context.Context, models.Grant) (models.Grant, error)
	Delete(context.Context, int64) error
}

type groupRepository interface {
	List(context.Context) ([]models.Group, error)
	Save(context.Context, models.Group) (models.Group, error)
	Delete(context.Context, int64) error
}

type rolloutRepository interface {
	ListPolicies(context.Context) ([]models.RolloutPolicy, error)
	GetPolicy(context.Context, int64) (models.RolloutPolicy, error)
	SavePolicy(context.Context, models.RolloutPolicy) (models.RolloutPolicy, error)
	DeletePolicy(context.Context, int64) error
	Get(context.Context, int64) (models.Rollout, error)
	Start(context.Context, models.Rollout) (models.Rollout, error)
	Save(context.Context, models.Rollout) (models.Rollout, error)
	CancelActive(context.Context, int64) error
	Ready(context.Context, time.Time) ([]int64, error)
}
type scopeRepository interface {
	Create(ctx context.Context, scope models.Scope) (models.Scope, error)
	Get(ctx context.Context, scopeID int64) (models.Scope, error)
	List(ctx context.Context, page models.PageRequest) (models.Page[models.Scope], error)
	ListVisible(ctx context.Context, page models.PageRequest, scopeIDs, keyIDs []int64) (models.Page[models.Scope], error)
	CanSee(ctx context.Context, scopeID int64, scopeIDs, keyIDs []int64) (bool, error)
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
	Get(ctx context.Context, keyID int64) (models.ConfigKey, error)
	GetDetails(ctx context.Context, keyID int64) (models.KeyDetails, error)
	List(ctx context.Context, filter models.KeyFilter) (models.Page[models.KeyDetails], error)
}
type snapshotRepository interface {
	Get(ctx context.Context) (models.Snapshot, error)
}
