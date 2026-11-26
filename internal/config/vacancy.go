package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type VacancyConfig struct {
	*PostgresConfig
	GRPCPort int
}

func LoadVacancy() (*VacancyConfig, error) {
	postgres, err := LoadPostgres()
	if err != nil {
		return nil, err
	}
	portStr := os.Getenv("VACANCY_GRPC_PORT")
	if portStr == "" {
		portStr = "2903"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid VACANCY_GRPC_PORT: %w", err)
	}
	if port <= 0 || port > 65535 {
		return nil, errors.New("Vacancy gRPC port must be between 1 and 65535")
	}
	return &VacancyConfig{
		PostgresConfig: postgres,
		GRPCPort:       port,
	}, nil
}
