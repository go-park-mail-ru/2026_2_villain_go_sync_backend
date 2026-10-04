package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ACCESS_TOKEN_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "168h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if string(cfg.JWTSecret) != "test-secret" {
		t.Error("JWTSecret differs from the supplied value")
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Errorf("AccessTokenTTL = %v, want 15m", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 168*time.Hour {
		t.Errorf("RefreshTokenTTL = %v, want 168h", cfg.RefreshTokenTTL)
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("ACCESS_TOKEN_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "168h")

	cfg, err := Load()
	if err == nil {
		t.Error("Load() returned no error for an empty JWT_SECRET")
	}
	if cfg != nil {
		t.Error("Load() returned config without JWT_SECRET")
	}
}

func TestLoad_InvalidTTL(t *testing.T) {
	tests := []struct {
		name       string
		accessTTL  string
		refreshTTL string
	}{
		{
			name:       "invalid access TTL",
			accessTTL:  "invalid",
			refreshTTL: "168h",
		},
		{
			name:       "invalid refresh TTL",
			accessTTL:  "15m",
			refreshTTL: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", "test-secret")
			t.Setenv("ACCESS_TOKEN_TTL", tt.accessTTL)
			t.Setenv("REFRESH_TOKEN_TTL", tt.refreshTTL)

			cfg, err := Load()
			if err == nil {
				t.Error("Load() returned no error for an invalid TTL")
			}
			if cfg != nil {
				t.Error("Load() returned config with an invalid TTL")
			}
		})
	}
}

func TestLoad_EmptyTTL(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ACCESS_TOKEN_TTL", "")
	t.Setenv("REFRESH_TOKEN_TTL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.AccessTokenTTL != 0 {
		t.Errorf("AccessTokenTTL = %v, want 0", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 0 {
		t.Errorf("RefreshTokenTTL = %v, want 0", cfg.RefreshTokenTTL)
	}
}
