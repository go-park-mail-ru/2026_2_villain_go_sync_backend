package onlineconf

import "testing"

func TestEvaluatePrecedence(t *testing.T) {
	include, exclude := true, false
	for _, tc := range []struct {
		name          string
		global        bool
		rollout       int16
		override      *bool
		wantEnabled   bool
		wantReason    Reason
		expectsBucket bool
	}{
		{"global disable beats include", false, 100, &include, false, ReasonGloballyDisabled, false},
		{"no override when disabled", false, 75, nil, false, ReasonGloballyDisabled, false},
		{"exclude beats 100 percent", true, 100, &exclude, false, ReasonForcedExclude, false},
		{"include beats zero percent", true, 0, &include, true, ReasonForcedInclude, false},
		{"zero percent", true, 0, nil, false, ReasonRolloutExcluded, true},
		{"one hundred percent", true, 100, nil, true, ReasonRolloutIncluded, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate(42, 7, tc.global, tc.rollout, tc.override)
			if result.Enabled != tc.wantEnabled || result.Reason != tc.wantReason ||
				(result.Bucket != nil) != tc.expectsBucket || result.RolloutPercentage != tc.rollout {
				t.Fatalf("Evaluate() = %+v", result)
			}
		})
	}
}

func TestEvaluateStableBuckets(t *testing.T) {
	for _, tc := range []struct {
		flag, user int64
		bucket     uint8
	}{
		{42, 7, 27},
		{42, 8, 23},
		{43, 7, 89},
	} {
		got := Evaluate(tc.flag, tc.user, true, 50, nil)
		if got.Bucket == nil || *got.Bucket != tc.bucket {
			t.Errorf("Evaluate(%d, %d) = %+v, want bucket %d", tc.flag, tc.user, got.Bucket, tc.bucket)
		}
	}
}

func TestIncreasingRolloutKeepsIncludedUsers(t *testing.T) {
	includedAt50 := 0
	for userID := int64(1); userID <= 1000; userID++ {
		at10 := Evaluate(9, userID, true, 10, nil)
		at50 := Evaluate(9, userID, true, 50, nil)
		repeat := Evaluate(9, userID, true, 50, nil)
		if (at10.Enabled && !at50.Enabled) || *at10.Bucket != *at50.Bucket || *at50.Bucket != *repeat.Bucket ||
			at50.Enabled != repeat.Enabled || *at50.Bucket >= 100 {
			t.Fatalf("unstable rollout for user %d", userID)
		}
		if at50.Enabled {
			includedAt50++
		}
	}
	if includedAt50 < 400 || includedAt50 > 600 {
		t.Fatalf("50 percent rollout included %d of 1000 users", includedAt50)
	}
}
