package config

import (
	"errors"
	"os"
)

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

func LoadS3() (*S3Config, error) {
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

	return &S3Config{
		Endpoint:  s3Endpoint,
		Region:    s3Region,
		Bucket:    s3Bucket,
		AccessKey: s3AccessKey,
		SecretKey: s3SecretKey,
	}, nil
}
