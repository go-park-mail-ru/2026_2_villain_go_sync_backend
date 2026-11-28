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
	AdminRoleOwner  AdminRole = "owner"
	AdminRoleAdmin  AdminRole = "admin"
	AdminRoleMember AdminRole = "member"
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
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ConfigKey struct {
	ID          int64
	ScopeID     int64
	ScopeName   string
	Name        string
	Description *string
	Type        KeyType
	OwnerUserID int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Config struct {
	ConfigKey
	Value json.RawMessage
}

type FeatureFlag struct {
	ConfigKey
	Enabled           bool
	RolloutPercentage int16
}

type FeatureFlagOverride struct {
	KeyID     int64
	UserID    int64
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Admin struct {
	UserID    int64
	Role      AdminRole
	GrantedBy *int64
	CreatedAt time.Time
}

type ChangeHistory struct {
	ID           int64
	ActorUserID  int64
	ResourceType ResourceType
	ResourceID   int64
	ScopeName    *string
	ResourceName string
	Action       Action
	OldValue     json.RawMessage
	NewValue     json.RawMessage
	CreatedAt    time.Time
}
