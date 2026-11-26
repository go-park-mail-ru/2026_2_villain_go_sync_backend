package profileimage

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
)

type ImageStorage interface {
	Upload(ctx context.Context, key string, body io.ReadSeeker, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
}

func Replace(
	ctx context.Context,
	storage ImageStorage,
	userID int,
	prefix string,
	oldKey *string,
	body io.ReadSeeker,
	contentType string,
	updateKey func(context.Context, int, *string) error,
) error {
	key := fmt.Sprintf("%s/%d/%s", prefix, userID, rand.Text())
	err := storage.Upload(ctx, key, body, contentType)
	if err != nil {
		return fmt.Errorf("upload profile image: %w", err)
	}
	err = updateKey(ctx, userID, &key)
	if err != nil {
		deleteErr := storage.Delete(ctx, key)
		if deleteErr != nil {
			return errors.Join(
				fmt.Errorf("update profile image key: %w", err),
				fmt.Errorf("delete profile image: %w", deleteErr),
			)
		}
		return fmt.Errorf("update profile image key: %w", err)
	}
	if oldKey != nil {
		err := storage.Delete(ctx, *oldKey)
		if err != nil {
			slog.Warn("failed to delete old profile image",
				"user_id", userID,
				"error", err,
			)
		}
	}
	return nil
}

func Delete(ctx context.Context, storage ImageStorage, userID int, oldKey *string, updateKey func(context.Context, int, *string) error) error {
	if oldKey == nil {
		return apperrors.ErrNotFound
	}

	err := updateKey(ctx, userID, nil)
	if err != nil {
		return fmt.Errorf("delete profile image key: %w", err)
	}
	err = storage.Delete(ctx, *oldKey)
	if err != nil {
		slog.Warn("failed to delete profile image",
			"user_id", userID,
			"error", err,
		)
	}
	return nil
}
