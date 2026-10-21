package auth

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
	roleKey
)

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func WithRole(ctx context.Context, role models.Role) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

func RoleFromContext(ctx context.Context) (models.Role, bool) {
	role, ok := ctx.Value(roleKey).(models.Role)
	return role, ok
}
