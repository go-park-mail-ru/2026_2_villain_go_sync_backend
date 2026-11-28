package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
)

func setConfig(t *testing.T, runtime, product string) {
	t.Helper()
	t.Setenv("ONLINECONF_DATABASE_URL", runtime)
	t.Setenv("ONLINECONF_PRODUCT_READER_DATABASE_URL", product)
	t.Setenv("ONLINECONF_SNAPSHOT_SECRET", "test-only-snapshot")
	t.Setenv("JWT_SECRET", "test-only-jwt")
}

func TestLoadConfig(t *testing.T) {
	setConfig(t, "runtime", "product")
	t.Setenv("ONLINECONF_ADDR", "")
	cfg, err := loadConfig()
	if err != nil || cfg.addr != ":8081" {
		t.Fatalf("default config: addr %q, %v", cfg.addr, err)
	}
	t.Setenv("ONLINECONF_ADDR", "127.0.0.1:9091")
	cfg, err = loadConfig()
	if err != nil || cfg.addr != "127.0.0.1:9091" {
		t.Fatalf("explicit address: addr %q, %v", cfg.addr, err)
	}
	for _, variable := range []string{
		"ONLINECONF_DATABASE_URL", "ONLINECONF_PRODUCT_READER_DATABASE_URL", "JWT_SECRET", "ONLINECONF_SNAPSHOT_SECRET",
	} {
		t.Run(variable, func(t *testing.T) {
			t.Setenv(variable, "")
			if _, err := loadConfig(); err == nil {
				t.Fatalf("missing %s accepted", variable)
			}
		})
	}
	t.Setenv("ONLINECONF_SNAPSHOT_SECRET", "test-only-jwt")
	if _, err := loadConfig(); err == nil {
		t.Fatal("shared JWT and snapshot secret accepted")
	}
}

func TestPostgresServiceStartAndShutdown(t *testing.T) {
	runtime, product := os.Getenv("ONLINECONF_TEST_DATABASE_URL"), os.Getenv("ONLINECONF_TEST_PRODUCT_READER_DATABASE_URL")
	if runtime == "" || product == "" {
		t.Skip("set both isolated OnlineConf HTTP test URLs")
	}
	setConfig(t, runtime, product)
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := probe.Addr().String()
	if err := probe.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ONLINECONF_ADDR", addr)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- run(ctx) }()

	client := &http.Client{Timeout: time.Second}
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-stopped:
			t.Fatalf("service exited before readiness: %v", err)
		case <-deadline.C:
			t.Fatal("service did not become healthy")
		case <-ticker.C:
			resp, err := client.Get("http://" + addr + "/health")
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("health status %d", resp.StatusCode)
			}
		}
		break
	}
	cancel()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("graceful shutdown timed out")
	}
}

type workerTestUsers struct{}

func (workerTestUsers) Exists(_ context.Context, userID int64) (bool, error) { return userID == 1, nil }

func TestPostgresWorkerAppliesApprovedStage(t *testing.T) {
	runtime := os.Getenv("ONLINECONF_TEST_DATABASE_URL")
	if runtime == "" {
		t.Skip("set the isolated, bootstrapped OnlineConf test URL")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	pool, err := openRuntimePool(ctx, runtime)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	u := usecase.New(repository.NewStore(pool), workerTestUsers{})
	const owner = 1
	scope, err := u.Scopes.Create(ctx, owner, models.Scope{Name: "worker_test"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := u.Scopes.Delete(ctx, owner, scope.ID); err != nil {
			t.Error(err)
		}
	}()
	flag, err := u.Flags.Create(ctx, owner, models.FeatureFlag{
		ConfigKey: models.ConfigKey{ScopeID: scope.ID, Name: "worker_flag", OwnerUserID: owner}, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := u.Flags.Delete(ctx, owner, flag.ID); err != nil {
			t.Error(err)
		}
	}()
	policy, err := u.Rollouts.SavePolicy(ctx, owner, models.RolloutPolicy{Name: "Worker test", Steps: []models.RolloutStep{
		{Percentage: 10}, {Percentage: 100, WaitSeconds: 1},
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := u.Rollouts.DeletePolicy(ctx, owner, policy.ID); err != nil {
			t.Error(err)
		}
	}()
	rollout, err := u.Rollouts.Start(ctx, owner, flag.ID, policy.ID)
	if err != nil {
		t.Fatal(err)
	}
	workerCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		runRollouts(workerCtx, u.Rollouts)
	}()
	defer func() {
		stop()
		<-done
	}()
	if _, err := u.Rollouts.Approve(ctx, owner, flag.ID, rollout.RunID, 0); err != nil {
		t.Fatal(err)
	}

	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Fatal("worker did not apply the approved step")
		case <-ticker.C:
			got, err := u.Flags.Get(ctx, owner, flag.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got.RolloutPercentage == 10 {
				current, err := u.Rollouts.Get(ctx, owner, flag.ID)
				if err != nil || current == nil || current.NextStep != 1 || current.ApprovedBy != nil {
					t.Fatalf("next stage = %+v, %v", current, err)
				}
				return
			}
		}
	}
}
