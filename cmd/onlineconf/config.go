package main

import (
	"errors"
	"os"
	"strings"
	"unicode"
)

type config struct {
	addr             string
	databaseURL      string
	productReaderURL string
	jwtSecret        string
	snapshotSecret   string
}

func loadConfig() (config, error) {
	cfg := config{
		addr:             os.Getenv("ONLINECONF_ADDR"),
		databaseURL:      os.Getenv("ONLINECONF_DATABASE_URL"),
		productReaderURL: os.Getenv("ONLINECONF_PRODUCT_READER_DATABASE_URL"),
		jwtSecret:        os.Getenv("JWT_SECRET"),
		snapshotSecret:   os.Getenv("ONLINECONF_SNAPSHOT_SECRET"),
	}
	if cfg.addr == "" {
		cfg.addr = ":8081"
	}
	switch {
	case cfg.databaseURL == "":
		return config{}, errors.New("ONLINECONF_DATABASE_URL is required")
	case cfg.productReaderURL == "":
		return config{}, errors.New("ONLINECONF_PRODUCT_READER_DATABASE_URL is required")
	case strings.TrimSpace(cfg.jwtSecret) == "":
		return config{}, errors.New("JWT_SECRET is required")
	case cfg.snapshotSecret == "" || strings.ContainsFunc(cfg.snapshotSecret, unicode.IsSpace):
		return config{}, errors.New("ONLINECONF_SNAPSHOT_SECRET must be a nonempty token without whitespace")
	case cfg.snapshotSecret == cfg.jwtSecret:
		return config{}, errors.New("ONLINECONF_SNAPSHOT_SECRET must differ from JWT_SECRET")
	}
	return cfg, nil
}
