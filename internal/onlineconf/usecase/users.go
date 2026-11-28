package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
)

type ProductUsers interface {
	Exists(context.Context, int64) (bool, error)
}

var ErrProductUsersUnavailable = errors.New("product users unavailable")

func requireUser(ctx context.Context, users ProductUsers, userID int64) error {
	if userID <= 0 {
		return apperrors.ValidationError("user id must be positive")
	}
	if users == nil {
		return ErrProductUsersUnavailable
	}
	exists, err := users.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProductUsersUnavailable, err)
	}
	if !exists {
		return apperrors.ValidationError("product user does not exist")
	}
	return nil
}
