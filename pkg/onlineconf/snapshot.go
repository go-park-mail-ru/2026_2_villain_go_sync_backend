package onlineconf

import (
	"bytes"
	"encoding/json"
)

type key struct {
	scope, name string
}

type flag struct {
	id      int64
	enabled bool
	rollout int16
}

type overrideKey struct {
	flagID, userID int64
}

type Snapshot struct {
	configs   map[key]json.RawMessage
	flags     map[key]flag
	overrides map[overrideKey]bool
}

type keyResponse struct {
	ScopeName string `json:"scope_name"`
	Name      string `json:"name"`
}

type configResponse struct {
	keyResponse
	Value json.RawMessage `json:"value"`
}

type flagResponse struct {
	keyResponse
	ID      int64  `json:"id"`
	Enabled *bool  `json:"enabled"`
	Rollout *int16 `json:"rollout_percentage"`
}

type overrideResponse struct {
	KeyID   int64 `json:"key_id"`
	UserID  int64 `json:"user_id"`
	Enabled *bool `json:"enabled"`
}

type snapshotResponse struct {
	Configs   []configResponse   `json:"configs"`
	Flags     []flagResponse     `json:"flags"`
	Overrides []overrideResponse `json:"overrides"`
}

func (r snapshotResponse) snapshot() (Snapshot, error) {
	if r.Configs == nil || r.Flags == nil || r.Overrides == nil {
		return Snapshot{}, ErrInvalidSnapshot
	}

	s := Snapshot{
		configs:   make(map[key]json.RawMessage, len(r.Configs)),
		flags:     make(map[key]flag, len(r.Flags)),
		overrides: make(map[overrideKey]bool, len(r.Overrides)),
	}
	for _, config := range r.Configs {
		k := key{config.ScopeName, config.Name}
		if k.scope == "" || k.name == "" || config.Value == nil || s.configs[k] != nil {
			return Snapshot{}, ErrInvalidSnapshot
		}
		s.configs[k] = config.Value
	}

	flagIDs := make(map[int64]bool, len(r.Flags))
	for _, item := range r.Flags {
		k := key{item.ScopeName, item.Name}
		if k.scope == "" || k.name == "" || item.ID <= 0 || flagIDs[item.ID] || s.flags[k].id != 0 ||
			s.configs[k] != nil || item.Enabled == nil || item.Rollout == nil || *item.Rollout < 0 || *item.Rollout > 100 {
			return Snapshot{}, ErrInvalidSnapshot
		}
		s.flags[k] = flag{item.ID, *item.Enabled, *item.Rollout}
		flagIDs[item.ID] = true
	}

	for _, override := range r.Overrides {
		k := overrideKey{override.KeyID, override.UserID}
		_, exists := s.overrides[k]
		if !flagIDs[override.KeyID] || override.UserID <= 0 || override.Enabled == nil || exists {
			return Snapshot{}, ErrInvalidSnapshot
		}
		s.overrides[k] = *override.Enabled
	}
	return s, nil
}

func (s Snapshot) JSON(scope, name string) (json.RawMessage, bool) {
	value, ok := s.configs[key{scope, name}]
	return bytes.Clone(value), ok
}

func (s Snapshot) Int(scope, name string, fallback int) int {
	return configValue(s, scope, name, fallback)
}

func (s Snapshot) Bool(scope, name string, fallback bool) bool {
	return configValue(s, scope, name, fallback)
}

func (s Snapshot) String(scope, name, fallback string) string {
	return configValue(s, scope, name, fallback)
}

func configValue[T int | bool | string](s Snapshot, scope, name string, fallback T) T {
	raw := s.configs[key{scope, name}]
	if raw == nil {
		return fallback
	}
	var value *T
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return fallback
	}
	return *value
}

func (s Snapshot) Evaluate(scope, name string, userID int64) (Evaluation, bool) {
	f, ok := s.flags[key{scope, name}]
	if !ok {
		return Evaluation{}, false
	}
	var override *bool
	if value, exists := s.overrides[overrideKey{f.id, userID}]; exists {
		override = &value
	}
	return Evaluate(f.id, userID, f.enabled, f.rollout, override), true
}

func (s Snapshot) Enabled(scope, name string, userID int64, fallback bool) bool {
	result, exists := s.Evaluate(scope, name, userID)
	if !exists {
		return fallback
	}
	return result.Enabled
}
