package onlineconf

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fixtureResponse struct {
	status int
	body   string
}

func waitUntil(t *testing.T, done func() bool) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(2 * time.Millisecond)
	defer ticker.Stop()
	for !done() {
		select {
		case <-ticker.C:
		case <-deadline.C:
			t.Fatal("snapshot condition was not met")
		}
	}
}

func stopCache(t *testing.T, cancel context.CancelFunc, cache *Cache) {
	t.Helper()
	cancel()
	select {
	case <-cache.Done():
	case <-time.After(2 * time.Second):
		t.Error("cache did not stop")
	}
}

func TestCacheInitialLoadRefreshRetainsAndRemoves(t *testing.T) {
	var current atomic.Value
	current.Store(fixtureResponse{200, snapshotBody("2", true)})
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := current.Load().(fixtureResponse)
		calls.Add(1)
		w.WriteHeader(response.status)
		fmt.Fprint(w, response.body)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cache := startCache(ctx, client, 10*time.Millisecond)
	defer stopCache(t, cancel, cache)

	loadedAt := cache.LastSuccessfulRefresh()
	if loadedAt.IsZero() {
		t.Fatal("initial refresh did not record its success")
	}
	original := cache.View()
	if original.Int("vacancy", "max_categories", 5) != 2 || !original.Enabled("vacancy", "creation_enabled", 3, false) {
		t.Fatal("initial snapshot was unavailable")
	}
	for _, response := range []fixtureResponse{
		{503, "offline"}, {200, `{"data":{"configs":[],"flags":[]}}`},
	} {
		current.Store(response)
		previous := calls.Load()
		waitUntil(t, func() bool { return calls.Load() >= previous+2 })
		if cache.View().Int("vacancy", "max_categories", 5) != 2 || cache.LastSuccessfulRefresh() != loadedAt {
			t.Fatal("failed refresh discarded the last good snapshot")
		}
	}
	current.Store(fixtureResponse{200, emptySnapshot})
	waitUntil(t, func() bool { _, ok := cache.View().JSON("vacancy", "max_categories"); return !ok })
	if cache.View().Int("vacancy", "max_categories", 5) != 5 ||
		cache.View().Enabled("vacancy", "creation_enabled", 3, false) ||
		original.Int("vacancy", "max_categories", 5) != 2 || !cache.LastSuccessfulRefresh().After(loadedAt) {
		t.Fatal("deleted data did not revert to defaults or previously acquired view was mutated")
	}
}

func TestCacheDefaultsUntilFirstSuccess(t *testing.T) {
	var available atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !available.Load() {
			w.WriteHeader(503)
			return
		}
		fmt.Fprint(w, snapshotBody("3", false))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cache := startCache(ctx, client, 10*time.Millisecond)
	defer stopCache(t, cancel, cache)
	view := cache.View()
	if view.Int("vacancy", "max_categories", 5) != 5 || !view.Enabled("vacancy", "creation_enabled", 42, true) {
		t.Fatal("unavailable OnlineConf did not use defaults")
	}
	available.Store(true)
	waitUntil(t, func() bool { return cache.View().Int("vacancy", "max_categories", 5) == 3 })
}

func TestCacheConcurrentReadersGetOneSnapshot(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1)%2 == 0 {
			fmt.Fprint(w, snapshotBody("2", false))
			return
		}
		fmt.Fprint(w, snapshotBody("1", true))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cache := startCache(ctx, client, 2*time.Millisecond)
	defer stopCache(t, cancel, cache)

	var bad atomic.Bool
	var readers sync.WaitGroup
	readCtx, stopReaders := context.WithCancel(ctx)
	for range 8 {
		readers.Go(func() {
			for readCtx.Err() == nil {
				view := cache.View()
				limit := view.Int("vacancy", "max_categories", 0)
				enabled := view.Enabled("vacancy", "creation_enabled", 7, false)
				if (limit == 1 && !enabled) || (limit == 2 && enabled) || (limit != 1 && limit != 2) {
					bad.Store(true)
					return
				}
			}
		})
	}
	defer func() {
		stopReaders()
		readers.Wait()
	}()
	waitUntil(t, func() bool { return calls.Load() >= 10 })
	stopReaders()
	readers.Wait()
	if bad.Load() {
		t.Fatal("reader observed a mixture of two snapshots")
	}
}
