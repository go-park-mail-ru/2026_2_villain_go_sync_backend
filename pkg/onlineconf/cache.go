package onlineconf

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

const RefreshInterval = 15 * time.Second

type generation struct {
	snapshot Snapshot
	loadedAt time.Time
}

type Cache struct {
	snapshot atomic.Pointer[generation]
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
				slog.WarnContext(ctx, "OnlineConf snapshot refresh failed", "error", err, "last_success", cache.LastSuccessfulRefresh())
			}
			return
		}
		cache.snapshot.Store(&generation{snapshot: snapshot, loadedAt: time.Now()})
	}

	refresh()
	go func() {
		defer close(cache.done)
		defer client.http.CloseIdleConnections()
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
		return snapshot.snapshot
	}
	return Snapshot{}
}

// LastSuccessfulRefresh is zero until a generation has been published.
func (c *Cache) LastSuccessfulRefresh() time.Time {
	if current := c.snapshot.Load(); current != nil {
		return current.loadedAt
	}
	return time.Time{}
}

func (c *Cache) Done() <-chan struct{} {
	return c.done
}
