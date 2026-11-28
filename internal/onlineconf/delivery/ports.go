package delivery

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

type Services struct {
	Scopes    scopeService
	Configs   configService
	Flags     featureFlagService
	Overrides overrideService
	Admins    adminService
	History   historyService
	Keys      keyService
	Snapshots snapshotService
}
type scopeService interface {
	Get(ctx context.Context, actorID, scopeID int64) (models.Scope, error)
	List(ctx context.Context, actorID int64, page models.PageRequest) (models.Page[models.Scope], error)
	Create(ctx context.Context, actorID int64, input models.Scope) (models.Scope, error)
	Update(ctx context.Context, actorID int64, input models.Scope) (models.Scope, error)
	Delete(ctx context.Context, actorID, scopeID int64) error
}
type configService interface {
	Get(ctx context.Context, actorID, keyID int64) (models.Config, error)
	Create(ctx context.Context, actorID int64, input models.Config) (models.Config, error)
	Update(ctx context.Context, actorID int64, input models.Config) (models.Config, error)
	Delete(ctx context.Context, actorID, keyID int64) error
}
type featureFlagService interface {
	Get(ctx context.Context, actorID, keyID int64) (models.FeatureFlag, error)
	Create(ctx context.Context, actorID int64, input models.FeatureFlag) (models.FeatureFlag, error)
	Update(ctx context.Context, actorID int64, input models.FeatureFlag) (models.FeatureFlag, error)
	SetEnabled(ctx context.Context, actorID, keyID int64, enabled bool) (models.FeatureFlag, error)
	SetRolloutPercentage(ctx context.Context, actorID, keyID int64, percentage int16) (models.FeatureFlag, error)
	Delete(ctx context.Context, actorID, keyID int64) error
	Evaluate(ctx context.Context, actorID, keyID, userID int64) (onlineconf.Evaluation, error)
}
type overrideService interface {
	ListByFlag(ctx context.Context, actorID, keyID int64,
		page models.PageRequest,
	) (models.Page[models.FeatureFlagOverride], error)
	ListByUser(ctx context.Context, actorID, userID int64,
		page models.PageRequest,
	) (models.Page[models.FeatureFlagOverride], error)
	Upsert(ctx context.Context, actorID int64, input models.FeatureFlagOverride) (models.FeatureFlagOverride, error)
	Delete(ctx context.Context, actorID, keyID, userID int64) error
}
type adminService interface {
	CheckAccess(ctx context.Context, actorID int64) (models.Admin, error)
	List(ctx context.Context, actorID int64, page models.PageRequest) (models.Page[models.Admin], error)
	Grant(ctx context.Context, actorID, userID int64, role models.AdminRole) (models.Admin, error)
	ChangeRole(ctx context.Context, actorID, userID int64, role models.AdminRole) (models.Admin, error)
	Revoke(ctx context.Context, actorID, userID int64) error
}
type historyService interface {
	Get(ctx context.Context, actorID, changeID int64) (models.ChangeHistory, error)
	List(ctx context.Context, actorID int64, filter models.HistoryFilter) (models.Page[models.ChangeHistory], error)
}
type keyService interface {
	Get(ctx context.Context, actorID, keyID int64) (models.KeyDetails, error)
	List(ctx context.Context, actorID int64, filter models.KeyFilter) (models.Page[models.KeyDetails], error)
}
type snapshotService interface {
	Get(ctx context.Context) (models.Snapshot, error)
}
