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

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func Load() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	accessTTL := time.Duration(0)
	if s := os.Getenv("ACCESS_TOKEN_TTL"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("parse ACCESS_TOKEN_TTL: %w", err)
		}
		accessTTL = d
	}

	refreshTTL := time.Duration(0)
	if s := os.Getenv("REFRESH_TOKEN_TTL"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("parse REFRESH_TOKEN_TTL: %w", err)
		}
		refreshTTL = d
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return nil, errors.New("DB_HOST is not set")
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return nil, errors.New("DB_USER is not set")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, errors.New("DB_PASSWORD is not set")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, errors.New("DB_NAME is not set")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		return nil, errors.New("S3_ENDPOINT is not set")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		return nil, errors.New("S3_REGION is not set")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		return nil, errors.New("S3_BUCKET is not set")
	}

	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	if s3AccessKey == "" {
		return nil, errors.New("S3_ACCESS_KEY is not set")
	}

	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	if s3SecretKey == "" {
		return nil, errors.New("S3_SECRET_KEY is not set")
	}

	return &Config{
		JWTSecret:       []byte(secret),
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBUser:          dbUser,
		DBPassword:      dbPassword,
		DBName:          dbName,
		S3Endpoint:      s3Endpoint,
		S3Region:        s3Region,
		S3Bucket:        s3Bucket,
		S3AccessKey:     s3AccessKey,
		S3SecretKey:     s3SecretKey,
	}, nil
}
