package delivery

import (
	"encoding/json"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
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

func (s scopeInput) validate(replace bool) error {
	if s.Name == "" {
		return apperrors.ValidationError("name is required")
	}
	if replace && !s.Description.Present {
		return apperrors.ValidationError("description is required on replacement; use null to clear")
	}
	return nil
}

type keyInput struct {
	scopeInput
	ScopeID     int64 `json:"scope_id"`
	OwnerUserID int64 `json:"owner_user_id"`
}

func (k keyInput) model(id int64, replace bool) (models.ConfigKey, error) {
	if err := k.scopeInput.validate(replace); err != nil {
		return models.ConfigKey{}, err
	}
	if k.ScopeID <= 0 || k.OwnerUserID <= 0 {
		return models.ConfigKey{}, apperrors.ValidationError("scope_id and owner_user_id must be positive")
	}
	return models.ConfigKey{
		ID: id, Name: k.Name, Description: k.Description.Value, ScopeID: k.ScopeID, OwnerUserID: k.OwnerUserID,
	}, nil
}

type configInput struct {
	keyInput
	Value json.RawMessage `json:"value"`
}

type enabledInput struct {
	Enabled field[bool] `json:"enabled"`
}

func (i enabledInput) value(required bool) (bool, error) {
	if i.Enabled.Value != nil {
		return *i.Enabled.Value, nil
	}
	if required || i.Enabled.Present {
		return false, apperrors.ValidationError("enabled must be a boolean")
	}
	return false, nil
}

type rolloutInput struct {
	RolloutPercentage field[int] `json:"rollout_percentage"`
}

func (i rolloutInput) value(required bool) (int16, error) {
	if i.RolloutPercentage.Value == nil && !required && !i.RolloutPercentage.Present {
		return 0, nil
	}
	if i.RolloutPercentage.Value == nil || *i.RolloutPercentage.Value < 0 || *i.RolloutPercentage.Value > 100 {
		return 0, apperrors.ValidationError("rollout_percentage must be an integer from 0 to 100")
	}
	return int16(*i.RolloutPercentage.Value), nil
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

func keyDetailsView(k repository.KeyDetails) keyDetailsDTO {
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

func pageView[T, R any](page repository.Page[T], convert func(T) R) pageDTO[R] {
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

func snapshotView(s repository.Snapshot) snapshotDTO {
	return snapshotDTO{mapItems(s.Configs, configView), mapItems(s.Flags, flagView), mapItems(s.Overrides, overrideView)}
}
