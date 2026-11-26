package models

import "testing"

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{"seeker", true},
		{"employer", true},
		{"admin", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run("role="+tt.role, func(t *testing.T) {
			if got := Role(tt.role).IsValid(); got != tt.want {
				t.Errorf("Role(%q).IsValid() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}
