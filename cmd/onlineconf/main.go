package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/delivery"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/productusers"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	startupTimeout  = 5 * time.Second
	headerTimeout   = 5 * time.Second
	readTimeout     = 15 * time.Second
	writeTimeout    = 20 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 15 * time.Second
	runtimePoolSize = 10
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("onlineconf stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	startup, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	pool, err := openRuntimePool(startup, cfg.databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	users, err := productusers.Open(startup, cfg.productReaderURL)
	if err != nil {
		return err
	}
	defer users.Close()
	cancel()

	health := func(ctx context.Context) error {
		if err := checkRuntimeSchema(ctx, pool); err != nil {
			return err
		}
		return users.Ping(ctx)
	}
	handler := newHTTPHandler(usecase.New(repository.NewStore(pool), users),
		auth.NewTokenManager([]byte(cfg.jwtSecret), 0, 0), cfg.snapshotSecret, health)
	server := &http.Server{
		Addr: cfg.addr, Handler: handler,
		ReadHeaderTimeout: headerTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
	stopped := make(chan error, 1)
	go func() {
		slog.Info("onlineconf HTTP server starting", "addr", cfg.addr)
		stopped <- server.ListenAndServe()
	}()

	select {
	case err := <-stopped:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			return errors.Join(err, server.Close())
		}
		if err := <-stopped; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}

func openRuntimePool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("invalid OnlineConf connection URL")
	}
	cfg.MaxConns = runtimePoolSize
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("cannot configure OnlineConf pool")
	}
	var valid bool
	err = pool.QueryRow(ctx, `SELECT current_database() = 'onlineconf'
	    AND current_user = 'onlineconf_runtime' AND session_user = current_user`).Scan(&valid)
	if err != nil || !valid {
		pool.Close()
		return nil, errors.New("could not connect to onlineconf as onlineconf_runtime")
	}
	if err := checkRuntimeSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, errors.New("OnlineConf schema is unavailable; verify migrations")
	}
	return pool, nil
}

func checkRuntimeSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "SELECT 1 FROM public.scope LIMIT 1")
	return err
}

func newHTTPHandler(u usecase.Usecases, tokens *auth.TokenManager, secret string, health func(context.Context) error) http.Handler {
	return delivery.New(delivery.Services{Scopes: u.Scopes, Configs: u.Configs, Flags: u.Flags, Overrides: u.Overrides, Admins: u.Admins, History: u.History, Keys: u.Keys, Snapshots: u.Snapshots}, tokens, secret, health)
}
