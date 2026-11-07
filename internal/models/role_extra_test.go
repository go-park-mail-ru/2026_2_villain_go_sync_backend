package models_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestRoleHelpers_Table(t *testing.T) {
	tests := []struct {
		name         string
		role         models.Role
		wantSeeker   bool
		wantEmployer bool
	}{
		{
			name:       "seeker",
			role:       models.RoleSeeker,
			wantSeeker: true,
		},
		{
			name:         "employer",
			role:         models.RoleEmployer,
			wantEmployer: true,
		},
		{
			name: "unknown role",
			role: models.Role("unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsSeeker(); got != tt.wantSeeker {
				t.Errorf("IsSeeker() = %v, want %v", got, tt.wantSeeker)
			}

			if got := tt.role.IsEmployer(); got != tt.wantEmployer {
				t.Errorf("IsEmployer() = %v, want %v", got, tt.wantEmployer)
			}
		})
	}
}
