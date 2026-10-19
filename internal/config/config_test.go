package config

import (
	"testing"
	"time"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ACCESS_TOKEN_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "168h")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "postgres")
	t.Setenv("DB_NAME", "villain_go")
}

func TestLoad(t *testing.T) {
	setValidEnv(t)

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

	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %q, want localhost", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %q, want 5432", cfg.DBPort)
	}
	if cfg.DBUser != "postgres" {
		t.Errorf("DBUser = %q, want postgres", cfg.DBUser)
	}
	if cfg.DBPassword != "postgres" {
		t.Errorf("DBPassword = %q, want postgres", cfg.DBPassword)
	}
	if cfg.DBName != "villain_go" {
		t.Errorf("DBName = %q, want villain_go", cfg.DBName)
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	setValidEnv(t)
	t.Setenv("JWT_SECRET", "")

	cfg, err := Load()
	if err == nil {
		t.Error("Load() returned no error for an empty JWT_SECRET")
	}
	if cfg != nil {
		t.Error("Load() returned config without JWT_SECRET")
	}
}

func TestLoad_MissingDBFields(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{"missing DB_HOST", "DB_HOST"},
		{"missing DB_USER", "DB_USER"},
		{"missing DB_PASSWORD", "DB_PASSWORD"},
		{"missing DB_NAME", "DB_NAME"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv(tt.env, "")

			cfg, err := Load()
			if err == nil {
				t.Errorf("Load() returned no error for empty %s", tt.env)
			}
			if cfg != nil {
				t.Errorf("Load() returned config with empty %s", tt.env)
			}
		})
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
			setValidEnv(t)
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
	setValidEnv(t)
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

func TestLoad_DefaultOptionalDBFields(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DB_PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}

	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %q, want default 5432", cfg.DBPort)
	}
}

func TestConfig_DSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBName:     "villain_go",
	}

	want := "postgres://postgres:postgres@localhost:5432/villain_go?sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}
