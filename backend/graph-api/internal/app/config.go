package app

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port int `env:"PORT" envDefault:"8080"`
	DB   DBConfig
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
