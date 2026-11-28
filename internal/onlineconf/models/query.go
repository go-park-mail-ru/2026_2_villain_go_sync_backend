package models

import "encoding/json"

const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

type KeyFilter struct {
	ScopeID  *int64
	Type     *KeyType
	Search   string
	Page     PageRequest
	Readable *ReadableKeys
}

type ReadableKeys struct {
	IDs          []int64
	ConfigScopes []int64
	FlagScopes   []int64
}

type KeyDetails struct {
	ConfigKey
	Value             json.RawMessage
	Enabled           *bool
	RolloutPercentage *int16
}

type Snapshot struct {
	Configs   []Config
	Flags     []FeatureFlag
	Overrides []FeatureFlagOverride
}

type PageRequest struct {
	Cursor *int64
	Limit  int
}

type Page[T any] struct {
	Items      []T
	NextCursor *int64
}

type HistoryFilter struct {
	ResourceType *ResourceType
	ResourceID   *int64
	ScopeName    *string
	Page         PageRequest
}
