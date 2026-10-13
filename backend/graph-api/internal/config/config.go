package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
)

type Log struct {
	Level       slog.Level `env:"LOG_LEVEL" envDefault:"info"`
	ServiceName string     `env:"SERVICE_NAME" envDefault:"graph-api"`
	Environment string     `env:"DEPLOYMENT_ENVIRONMENT" envDefault:"local"`
}

type Config struct {
	Port int `env:"PORT" envDefault:"8080"`
	DB   DBConfig
	Log  Log
}

type DBConfig struct {
	URL string `env:"DATABASE_URL,required"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil,
			fmt.Errorf("parse app: %w", err)
	}
	return cfg, nil
}
