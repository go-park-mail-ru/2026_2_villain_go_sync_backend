package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type SeekerConfig struct {
	*PostgresConfig
	GRPCPort int
}

func LoadSeeker() (*SeekerConfig, error) {
	postgres, err := LoadPostgres()
	if err != nil {
		return nil, err
	}
	portStr := os.Getenv("SEEKER_GRPC_PORT")
	if portStr == "" {
		portStr = "2007"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SEEKER_GRPC_PORT: %w", err)
	}
	if port <= 0 || port > 65535 {
		return nil, errors.New("Seeker gRPC port must be between 1 and 65535")
	}
	return &SeekerConfig{
		PostgresConfig: postgres,
		GRPCPort:       port,
	}, nil
}
