package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/server"
)

const shutdownTimeout = 15 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		slog.Error("backend stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	const addr = ":8080"
	handler, closeResources, err := server.New(ctx)
	if err != nil {
		return err
	}
	defer closeResources()

	requests := &activeRequests{next: handler}
	defer requests.wait()
	srv := &http.Server{Addr: addr, Handler: requests}
	stopped := make(chan error, 1)
	go func() {
		slog.Info("backend starting", "address", addr)
		stopped <- srv.ListenAndServe()
	}()

	select {
	case err := <-stopped:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			return errors.Join(err, srv.Close())
		}
		if err := <-stopped; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}

// After a forced close, canceled HTTP handlers may still be releasing resources.
type activeRequests struct {
	next     http.Handler
	mu       sync.Mutex
	stopping bool
	running  sync.WaitGroup
}

func (a *activeRequests) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	if a.stopping {
		a.mu.Unlock()
		return
	}
	a.running.Add(1)
	a.mu.Unlock()
	defer a.running.Done()
	a.next.ServeHTTP(w, r)
}

func (a *activeRequests) wait() {
	a.mu.Lock()
	a.stopping = true
	a.mu.Unlock()
	a.running.Wait()
}
