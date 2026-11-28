package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
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
