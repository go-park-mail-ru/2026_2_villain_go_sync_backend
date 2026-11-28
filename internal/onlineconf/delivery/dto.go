package delivery

import (
	"encoding/json"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

type field[T any] struct {
	Present bool
	Value   *T
}

func (f *field[T]) UnmarshalJSON(raw []byte) error {
	f.Present = true
	return json.Unmarshal(raw, &f.Value)
}

type scopeInput struct {
	Name        string        `json:"name"`
	Description field[string] `json:"description"`
}

type keyInput struct {
	scopeInput
	ScopeID     int64 `json:"scope_id"`
	OwnerUserID int64 `json:"owner_user_id"`
}

type configInput struct {
	keyInput
	Value json.RawMessage `json:"value"`
}

type enabledInput struct {
	Enabled field[bool] `json:"enabled"`
}

type rolloutInput struct {
	RolloutPercentage field[int] `json:"rollout_percentage"`
}

type flagInput struct {
	keyInput
	enabledInput
	rolloutInput
}

type roleInput struct {
	Role models.AdminRole `json:"role"`
}

type grantInput struct {
	roleInput
	UserID int64 `json:"user_id"`
}

type scopeDTO struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func scopeView(s models.Scope) scopeDTO {
	return scopeDTO{s.ID, s.Name, s.Description, s.CreatedAt, s.UpdatedAt}
}

type keyDTO struct {
	ID          int64          `json:"id"`
	ScopeID     int64          `json:"scope_id"`
	ScopeName   string         `json:"scope_name"`
	Name        string         `json:"name"`
	FullName    string         `json:"full_name"`
	Description *string        `json:"description"`
	Type        models.KeyType `json:"type"`
	OwnerUserID int64          `json:"owner_user_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func keyView(k models.ConfigKey) keyDTO {
	return keyDTO{k.ID, k.ScopeID, k.ScopeName, k.Name, k.ScopeName + "." + k.Name,
		k.Description, k.Type, k.OwnerUserID, k.CreatedAt, k.UpdatedAt}
}

type keyDetailsDTO struct {
	keyDTO
	Value             json.RawMessage `json:"value,omitempty"`
	Enabled           *bool           `json:"enabled,omitempty"`
	RolloutPercentage *int16          `json:"rollout_percentage,omitempty"`
}

func keyDetailsView(k models.KeyDetails) keyDetailsDTO {
	return keyDetailsDTO{keyView(k.ConfigKey), k.Value, k.Enabled, k.RolloutPercentage}
}

type configDTO struct {
	keyDTO
	Value json.RawMessage `json:"value"`
}

func configView(c models.Config) configDTO {
	return configDTO{keyView(c.ConfigKey), c.Value}
}

type flagDTO struct {
	keyDTO
	Enabled           bool  `json:"enabled"`
	RolloutPercentage int16 `json:"rollout_percentage"`
}

func flagView(f models.FeatureFlag) flagDTO {
	return flagDTO{keyView(f.ConfigKey), f.Enabled, f.RolloutPercentage}
}

type overrideDTO struct {
	KeyID     int64     `json:"key_id"`
	UserID    int64     `json:"user_id"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func overrideView(o models.FeatureFlagOverride) overrideDTO {
	return overrideDTO{o.KeyID, o.UserID, o.Enabled, o.CreatedAt, o.UpdatedAt}
}

type adminDTO struct {
	UserID    int64            `json:"user_id"`
	Role      models.AdminRole `json:"role"`
	GrantedBy *int64           `json:"granted_by"`
	CreatedAt time.Time        `json:"created_at"`
}

func adminView(a models.Admin) adminDTO {
	return adminDTO{a.UserID, a.Role, a.GrantedBy, a.CreatedAt}
}

type historyDTO struct {
	ID           int64               `json:"id"`
	ActorUserID  int64               `json:"actor_user_id"`
	ResourceType models.ResourceType `json:"resource_type"`
	ResourceID   int64               `json:"resource_id"`
	ScopeName    *string             `json:"scope_name"`
	ResourceName string              `json:"resource_name"`
	Action       models.Action       `json:"action"`
	OldValue     json.RawMessage     `json:"old_value,omitempty"`
	NewValue     json.RawMessage     `json:"new_value,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}

func historyView(h models.ChangeHistory) historyDTO {
	return historyDTO{h.ID, h.ActorUserID, h.ResourceType, h.ResourceID, h.ScopeName, h.ResourceName,
		h.Action, h.OldValue, h.NewValue, h.CreatedAt}
}

type pageDTO[T any] struct {
	Items      []T    `json:"items"`
	NextCursor *int64 `json:"next_cursor"`
}

func pageView[T, R any](page models.Page[T], convert func(T) R) pageDTO[R] {
	return pageDTO[R]{mapItems(page.Items, convert), page.NextCursor}
}

func mapItems[T, R any](items []T, convert func(T) R) []R {
	result := make([]R, len(items))
	for i, item := range items {
		result[i] = convert(item)
	}
	return result
}

type snapshotDTO struct {
	Configs   []configDTO   `json:"configs"`
	Flags     []flagDTO     `json:"flags"`
	Overrides []overrideDTO `json:"overrides"`
}

func snapshotView(s models.Snapshot) snapshotDTO {
	return snapshotDTO{mapItems(s.Configs, configView), mapItems(s.Flags, flagView), mapItems(s.Overrides, overrideView)}
}

type evaluationDTO struct {
	Enabled           bool              `json:"enabled"`
	Reason            onlineconf.Reason `json:"reason"`
	RolloutPercentage int16             `json:"rollout_percentage"`
	Bucket            *uint8            `json:"bucket,omitempty"`
}
