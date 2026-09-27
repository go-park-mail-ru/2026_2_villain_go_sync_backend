package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	JWTSecret       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	accessTTL := 15 * time.Minute
	if s := os.Getenv("ACCESS_TOKEN_TTL"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("parse ACCESS_TOKEN_TTL: %w", err)
		}
		accessTTL = d
	}

	refreshTTL := 168 * time.Hour
	if s := os.Getenv("REFRESH_TOKEN_TTL"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("parse REFRESH_TOKEN_TTL: %w", err)
		}
		refreshTTL = d
	}

	return &Config{
		JWTSecret:       []byte(secret),
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}, nil
}
