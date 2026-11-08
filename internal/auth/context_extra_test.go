package auth_test

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestUserIDContext_Table(t *testing.T) {
	tests := []struct {
		name   string
		withID bool
		wantID int64
		wantOK bool
	}{
		{
			name:   "context without user ID",
			wantOK: false,
		},
		{
			name:   "context with user ID",
			withID: true,
			wantID: 42,
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.withID {
				ctx = auth.WithUserID(ctx, 42)
			}

			id, ok := auth.UserIDFromContext(ctx)

			if ok != tt.wantOK {
				t.Errorf("expected ok=%v, got %v", tt.wantOK, ok)
			}

			if ok && id != tt.wantID {
				t.Errorf("expected ID=%d, got %d", tt.wantID, id)
			}
		})
	}
}

func TestRoleContext_Table(t *testing.T) {
	tests := []struct {
		name   string
		role   models.Role
		set    bool
		wantOK bool
	}{
		{
			name: "context without role",
		},
		{
			name:   "seeker role",
			role:   models.RoleSeeker,
			set:    true,
			wantOK: true,
		},
		{
			name:   "employer role",
			role:   models.RoleEmployer,
			set:    true,
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.set {
				ctx = auth.WithRole(ctx, tt.role)
			}

			role, ok := auth.RoleFromContext(ctx)

			if ok != tt.wantOK {
				t.Errorf("expected ok=%v, got %v", tt.wantOK, ok)
			}

			if ok && role != tt.role {
				t.Errorf("expected role=%s, got %s", tt.role, role)
			}
		})
	}
}
