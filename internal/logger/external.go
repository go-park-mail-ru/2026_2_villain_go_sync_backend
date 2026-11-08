package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/requestid"
)

type SourceType string

const (
	SourcePostgres SourceType = "postgres"
	SourceS3       SourceType = "s3"
)

func LogExternalCall(ctx context.Context, source SourceType, method string, duration time.Duration, err error) {
	args := []any{
		"request_id", requestid.FromContext(ctx),
		"service", source,
		"method", method,
		"duration_ms", duration.Milliseconds(),
	}

	if err != nil {
		args = append(args, "status", "error", "error", err.Error())
		slog.ErrorContext(ctx, "external call failed", args...)
	} else {
		args = append(args, "status", "success")
		slog.InfoContext(ctx, "external call", args...)
	}
}
