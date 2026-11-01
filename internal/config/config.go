package config

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	Env         string
	LogLevel    slog.Level
	AppVersion  string
	GRPCAddr    string
	HTTPAddr    string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		Env:         cmp.Or(os.Getenv("ENV"), "development"),
		AppVersion:  cmp.Or(os.Getenv("APP_VERSION"), "0.0.0-dev"),
		GRPCAddr:    cmp.Or(os.Getenv("GRPC_ADDR"), ":9090"),
		HTTPAddr:    cmp.Or(os.Getenv("HTTP_ADDR"), ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("config: DATABASE_URL is required")
	}

	logLevel := cmp.Or(os.Getenv("LOG_LEVEL"), "info")
	if err := cfg.LogLevel.UnmarshalText([]byte(logLevel)); err != nil {
		return Config{}, fmt.Errorf("config: LOG_LEVEL=%q is invalid: %w", logLevel, err)
	}
	return cfg, nil
}
