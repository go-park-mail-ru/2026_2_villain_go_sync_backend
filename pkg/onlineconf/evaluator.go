package onlineconf

import (
	"crypto/sha256"
	"encoding/binary"
)

const MaxRolloutPercentage = 100

type Reason string

const (
	ReasonGloballyDisabled Reason = "globally_disabled"
	ReasonForcedInclude    Reason = "forced_include"
	ReasonForcedExclude    Reason = "forced_exclude"
	ReasonRolloutIncluded  Reason = "rollout_included"
	ReasonRolloutExcluded  Reason = "rollout_excluded"
)

type Evaluation struct {
	Enabled           bool   `json:"enabled"`
	Reason            Reason `json:"reason"`
	RolloutPercentage int16  `json:"rollout_percentage"`
	Bucket            *uint8 `json:"bucket,omitempty"`
}

func Evaluate(flagID, userID int64, enabled bool, rolloutPercentage int16, override *bool) Evaluation {
	result := Evaluation{RolloutPercentage: rolloutPercentage}
	switch {
	case !enabled:
		result.Reason = ReasonGloballyDisabled
	case override != nil && !*override:
		result.Reason = ReasonForcedExclude
	case override != nil:
		result.Enabled = true
		result.Reason = ReasonForcedInclude
	default:
		var key [16]byte
		binary.BigEndian.PutUint64(key[:8], uint64(flagID))
		binary.BigEndian.PutUint64(key[8:], uint64(userID))
		digest := sha256.Sum256(key[:])
		bucket := uint8(binary.BigEndian.Uint64(digest[:8]) % MaxRolloutPercentage)
		result.Bucket = &bucket
		result.Enabled = int16(bucket) < rolloutPercentage
		result.Reason = ReasonRolloutExcluded
		if result.Enabled {
			result.Reason = ReasonRolloutIncluded
		}
	}
	return result
}
