package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type SnapshotRepository struct {
	db database
}

type Snapshot struct {
	Configs   []models.Config
	Flags     []models.FeatureFlag
	Overrides []models.FeatureFlagOverride
}

func (r SnapshotRepository) Get(ctx context.Context) (Snapshot, error) {
	rows, err := r.db.Query(ctx, "SELECT "+keyColumns+`,
		c.value, f.enabled, f.rollout_percentage, o.user_id, o.enabled, o.created_at, o.updated_at`+keyFrom+`
		LEFT JOIN public.config c ON c.key_id = k.key_id
		LEFT JOIN public.feature_flag f ON f.key_id = k.key_id
		LEFT JOIN public.feature_flag_override o ON o.key_id = f.key_id
		ORDER BY k.key_id, o.user_id`)
	if err != nil {
		return Snapshot{}, classifyError(ctx, err)
	}
	defer rows.Close()

	snapshot := Snapshot{
		Configs:   []models.Config{},
		Flags:     []models.FeatureFlag{},
		Overrides: []models.FeatureFlagOverride{},
	}
	var lastFlagID int64
	hasLastFlag := false
	for rows.Next() {
		var key models.ConfigKey
		var value json.RawMessage
		var flagEnabled, overrideEnabled *bool
		var rollout *int16
		var userID *int64
		var createdAt, updatedAt *time.Time
		err := rows.Scan(append(keyFields(&key), &value, &flagEnabled, &rollout, &userID, &overrideEnabled, &createdAt, &updatedAt)...)
		if err != nil {
			return Snapshot{}, classifyError(ctx, err)
		}

		if key.Type == models.KeyTypeConfig && value != nil {
			snapshot.Configs = append(snapshot.Configs, models.Config{ConfigKey: key, Value: value})
			continue
		}
		if key.Type != models.KeyTypeFeatureFlag || flagEnabled == nil || rollout == nil {
			return Snapshot{}, fmt.Errorf("config key %d has no matching subtype", key.ID)
		}
		if !hasLastFlag || lastFlagID != key.ID {
			snapshot.Flags = append(snapshot.Flags, models.FeatureFlag{
				ConfigKey: key, Enabled: *flagEnabled, RolloutPercentage: *rollout,
			})
			lastFlagID, hasLastFlag = key.ID, true
		}
		if userID != nil {
			snapshot.Overrides = append(snapshot.Overrides, models.FeatureFlagOverride{
				KeyID: key.ID, UserID: *userID, Enabled: *overrideEnabled, CreatedAt: *createdAt, UpdatedAt: *updatedAt,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, classifyError(ctx, err)
	}
	return snapshot, nil
}
