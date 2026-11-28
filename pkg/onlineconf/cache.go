package onlineconf

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

const RefreshInterval = 15 * time.Second

type Cache struct {
	snapshot atomic.Pointer[Snapshot]
	done     chan struct{}
}

func StartCache(ctx context.Context, client *Client) *Cache {
	return startCache(ctx, client, RefreshInterval)
}

func startCache(ctx context.Context, client *Client, interval time.Duration) *Cache {
	cache := &Cache{done: make(chan struct{})}
	refresh := func() {
		snapshot, err := client.FetchSnapshot(ctx)
		if err != nil {
			if ctx.Err() == nil {
				slog.WarnContext(ctx, "OnlineConf snapshot refresh failed", "error", err)
			}
			return
		}
		cache.snapshot.Store(&snapshot)
	}

	refresh()
	go func() {
		defer close(cache.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refresh()
			}
		}
	}()
	return cache
}

func (c *Cache) View() Snapshot {
	if snapshot := c.snapshot.Load(); snapshot != nil {
		return *snapshot
	}
	return Snapshot{}
}

func (c *Cache) Done() <-chan struct{} {
	return c.done
}
