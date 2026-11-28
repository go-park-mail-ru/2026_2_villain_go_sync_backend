package usecase

import (
	"encoding/json"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type flagDeletion struct {
	flag      models.FeatureFlag
	overrides []models.FeatureFlagOverride
}

// History keeps the field names already exposed inside old_value and new_value.
// On read, historical objects are returned as stored, including bootstrap records.
func marshalHistory(value any) (json.RawMessage, error) {
	var fields map[string]any
	switch v := value.(type) {
	case models.Scope:
		fields = map[string]any{"ID": v.ID, "name": v.Name, "description": v.Description, "created_at": v.CreatedAt, "updated_at": v.UpdatedAt}
	case models.Config:
		fields = historyKey(v.ConfigKey)
		fields["value"] = v.Value
	case models.FeatureFlag:
		fields = historyFlag(v)
	case flagDeletion:
		fields = historyFlag(v.flag)
		overrides := make([]map[string]any, 0, len(v.overrides))
		for _, override := range v.overrides {
			overrides = append(overrides, historyOverride(override))
		}
		fields["overrides"] = overrides
	case models.FeatureFlagOverride:
		fields = historyOverride(v)
	case models.Admin:
		fields = map[string]any{"UserID": v.UserID, "role": v.Role, "GrantedBy": v.GrantedBy, "created_at": v.CreatedAt}
	default:
		return nil, fmt.Errorf("unsupported audit snapshot %T", value)
	}
	return json.Marshal(fields)
}

func historyKey(k models.ConfigKey) map[string]any {
	return map[string]any{
		"ID": k.ID, "ScopeID": k.ScopeID, "scope_name": k.ScopeName, "name": k.Name, "description": k.Description,
		"type": k.Type, "OwnerUserID": k.OwnerUserID, "created_at": k.CreatedAt, "updated_at": k.UpdatedAt,
	}
}

func historyFlag(f models.FeatureFlag) map[string]any {
	fields := historyKey(f.ConfigKey)
	fields["enabled"] = f.Enabled
	fields["RolloutPercentage"] = f.RolloutPercentage
	return fields
}

func historyOverride(o models.FeatureFlagOverride) map[string]any {
	return map[string]any{"KeyID": o.KeyID, "UserID": o.UserID, "enabled": o.Enabled, "created_at": o.CreatedAt, "updated_at": o.UpdatedAt}
}
