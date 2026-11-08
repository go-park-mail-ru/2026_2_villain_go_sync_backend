package s3

import (
	"context"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/logger"
)

type LoggingAdapter struct {
	next *Adapter
}

func NewLoggingAdapter(next *Adapter) *LoggingAdapter {
	return &LoggingAdapter{next: next}
}

func (l *LoggingAdapter) Upload(ctx context.Context, key string, body io.ReadSeeker, contentType string) error {
	start := time.Now()
	err := l.next.Upload(ctx, key, body, contentType)
	logger.LogExternalCall(ctx, "s3", "upload", start, err)
	return err
}

func (l *LoggingAdapter) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	start := time.Now()
	body, contentType, err := l.next.Download(ctx, key)
	logger.LogExternalCall(ctx, "s3", "download", start, err)
	return body, contentType, err
}

func (l *LoggingAdapter) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := l.next.Delete(ctx, key)
	logger.LogExternalCall(ctx, "s3", "delete", start, err)
	return err
}
