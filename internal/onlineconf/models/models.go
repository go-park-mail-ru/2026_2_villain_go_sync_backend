package models

import (
	"encoding/json"
	"time"
)

type KeyType string

const (
	KeyTypeConfig      KeyType = "config"
	KeyTypeFeatureFlag KeyType = "feature_flag"
)

type AdminRole string

const (
	AdminRoleOwner AdminRole = "owner"
	AdminRoleAdmin AdminRole = "admin"
)

type ResourceType string

const (
	ResourceTypeScope               ResourceType = "scope"
	ResourceTypeConfig              ResourceType = "config"
	ResourceTypeFeatureFlag         ResourceType = "feature_flag"
	ResourceTypeFeatureFlagOverride ResourceType = "feature_flag_override"
	ResourceTypeAdmin               ResourceType = "admin"
)

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Scope struct {
	ID          int64
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ConfigKey struct {
	ID          int64
	ScopeID     int64
	ScopeName   string  `json:"scope_name"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Type        KeyType `json:"type"`
	OwnerUserID int64
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Config struct {
	ConfigKey
	Value json.RawMessage `json:"value"`
}

type FeatureFlag struct {
	ConfigKey
	Enabled           bool `json:"enabled"`
	RolloutPercentage int16
}

type FeatureFlagOverride struct {
	KeyID     int64
	UserID    int64
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Admin struct {
	UserID    int64
	Role      AdminRole `json:"role"`
	GrantedBy *int64
	CreatedAt time.Time `json:"created_at"`
}

type ChangeHistory struct {
	ID           int64
	ActorUserID  int64
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   int64
	ScopeName    *string         `json:"scope_name"`
	ResourceName string          `json:"resource_name"`
	Action       Action          `json:"action"`
	OldValue     json.RawMessage `json:"old_value"`
	NewValue     json.RawMessage `json:"new_value"`
	CreatedAt    time.Time       `json:"created_at"`
}
