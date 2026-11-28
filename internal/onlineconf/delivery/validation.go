package delivery

import (
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

func (s scopeInput) validate(replace bool) error {
	if s.Name == "" {
		return apperrors.ValidationError("name is required")
	}
	if replace && !s.Description.Present {
		return apperrors.ValidationError("description is required on replacement; use null to clear")
	}
	return nil
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

func (i enabledInput) value(required bool) (bool, error) {
	if i.Enabled.Value != nil {
		return *i.Enabled.Value, nil
	}
	if required || i.Enabled.Present {
		return false, apperrors.ValidationError("enabled must be a boolean")
	}
	return false, nil
}

func (i rolloutInput) value(required bool) (int16, error) {
	if i.RolloutPercentage.Value == nil && !required && !i.RolloutPercentage.Present {
		return 0, nil
	}
	if i.RolloutPercentage.Value == nil || *i.RolloutPercentage.Value < 0 || *i.RolloutPercentage.Value > onlineconf.MaxRolloutPercentage {
		return 0, apperrors.ValidationError(fmt.Sprintf("rollout_percentage must be an integer from 0 to %d", onlineconf.MaxRolloutPercentage))
	}
	return int16(*i.RolloutPercentage.Value), nil
}
