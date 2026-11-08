package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/requestid"
)

func LogExternalCall(ctx context.Context, service, method string, start time.Time, err error) {
	args := []any{
		"request_id", requestid.FromContext(ctx),
		"service", service,
		"method", method,
		"duration_ms", time.Since(start).Milliseconds(),
	}

	if err != nil {
		args = append(args, "status", "error", "error", err.Error())
		slog.ErrorContext(ctx, "external call failed", args...)
	} else {
		args = append(args, "status", "success")
		slog.InfoContext(ctx, "external call", args...)
	}
}
