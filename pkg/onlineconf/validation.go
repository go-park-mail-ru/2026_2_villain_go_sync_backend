package onlineconf

import "encoding/json"

func (r snapshotResponse) snapshot() (Snapshot, error) {
	if r.Configs == nil || r.Flags == nil || r.Overrides == nil {
		return Snapshot{}, ErrInvalidSnapshot
	}

	s := Snapshot{
		integers:  make(map[key]int),
		configs:   make(map[key]json.RawMessage, len(r.Configs)),
		flags:     make(map[key]flag, len(r.Flags)),
		overrides: make(map[overrideKey]bool, len(r.Overrides)),
	}
	if err := s.addConfigs(r.Configs); err != nil {
		return Snapshot{}, err
	}
	ids, err := s.addFlags(r.Flags)
	if err != nil {
		return Snapshot{}, err
	}
	if err := s.addOverrides(r.Overrides, ids); err != nil {
		return Snapshot{}, err
	}
	return s, nil
}
func (s Snapshot) addConfigs(configs []configResponse) error {
	for _, config := range configs {
		k := key{config.ScopeName, config.Name}
		if k.scope == "" || k.name == "" || config.Value == nil || s.configs[k] != nil {
			return ErrInvalidSnapshot
		}
		s.configs[k] = config.Value
		var integer *int
		if json.Unmarshal(config.Value, &integer) == nil && integer != nil {
			s.integers[k] = *integer
		}
	}

	return nil
}
func (s Snapshot) addFlags(flags []flagResponse) (map[int64]bool, error) {
	flagIDs := make(map[int64]bool, len(flags))
	for _, item := range flags {
		k := key{item.ScopeName, item.Name}
		if k.scope == "" || k.name == "" || item.ID <= 0 || flagIDs[item.ID] || s.flags[k].id != 0 ||
			s.configs[k] != nil || item.Enabled == nil || item.Rollout == nil || *item.Rollout < 0 || *item.Rollout > MaxRolloutPercentage {
			return nil, ErrInvalidSnapshot
		}
		s.flags[k] = flag{item.ID, *item.Enabled, *item.Rollout}
		flagIDs[item.ID] = true
	}

	return flagIDs, nil
}
func (s Snapshot) addOverrides(overrides []overrideResponse, flagIDs map[int64]bool) error {
	for _, override := range overrides {
		k := overrideKey{override.KeyID, override.UserID}
		_, exists := s.overrides[k]
		if !flagIDs[override.KeyID] || override.UserID <= 0 || override.Enabled == nil || exists {
			return ErrInvalidSnapshot
		}
		s.overrides[k] = *override.Enabled
	}
	return nil
}
