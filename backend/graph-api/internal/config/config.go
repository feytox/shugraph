package config

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
)

type Log struct {
	Level       slog.Level `env:"LOG_LEVEL" envDefault:"info"`
	ServiceName string     `env:"SERVICE_NAME" envDefault:"graph-api"`
	Environment string     `env:"DEPLOYMENT_ENVIRONMENT" envDefault:"local"`
}

type Serve struct {
	Port int `env:"PORT" envDefault:"8080"`
	DB   DBConfig
	Log  Log
}

type DBConfig struct {
	URL string `env:"DATABASE_URL,required"`
}

func Load[T any]() (T, error) {
	return env.ParseAs[T]()
}
