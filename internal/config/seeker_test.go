package config

import "testing"

func TestLoadSeeker(t *testing.T) {
	for _, tc := range []struct {
		name  string
		env   string
		port  int
		badDB bool
	}{
		{name: "default", port: 2007},
		{name: "override", env: "3904", port: 3904},
		{name: "non-numeric port", env: "invalid"},
		{name: "zero port", env: "0"},
		{name: "port too high", env: "65536"},
		{name: "missing database host", badDB: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv("JWT_SECRET", "")
			t.Setenv("SEEKER_GRPC_PORT", tc.env)
			if tc.badDB {
				t.Setenv("DB_HOST", "")
			}
			cfg, err := LoadSeeker()
			if tc.port == 0 {
				if err == nil || cfg != nil {
					t.Fatalf("expected configuration error, got %v", err)
				}
				return
			}
			if err != nil || cfg == nil || cfg.GRPCPort != tc.port || cfg.Host != "localhost" {
				t.Fatalf("invalid seeker configuration for port %d: %v", tc.port, err)
			}
		})
	}
}
