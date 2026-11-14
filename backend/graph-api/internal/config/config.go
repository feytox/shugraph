package config

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type Log struct {
	Level       slog.Level `env:"LOG_LEVEL" envDefault:"info"`
	ServiceName string     `env:"SERVICE_NAME" envDefault:"graph-api"`
	Environment string     `env:"DEPLOYMENT_ENVIRONMENT" envDefault:"local"`
}

type Database struct {
	URL             string        `env:"DATABASE_URL,required,notEmpty"`
	MaxConns        int32         `env:"DB_MAX_CONNS" envDefault:"10"`
	MinConns        int32         `env:"DB_MIN_CONNS" envDefault:"1"`
	MaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"30m"`
	ConnectTimeout  time.Duration `env:"DB_CONNECT_TIMEOUT" envDefault:"5s"`
}

type HTTP struct {
	Port              int           `env:"PORT" envDefault:"8080"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

type Groups struct {
	Window                  time.Duration `env:"GROUP_WINDOW" envDefault:"24h"`
	FarmMinClients          int           `env:"FARM_MIN_CLIENTS" envDefault:"3"`
	FarmMinCollectorSenders int           `env:"FARM_MIN_COLLECTOR_SENDERS" envDefault:"2"`
}

func (g Groups) Rules() domain.GroupRules {
	return domain.GroupRules{
		Window: g.Window,
		Farm: domain.FarmRules{
			MinClients:          g.FarmMinClients,
			MinCollectorSenders: g.FarmMinCollectorSenders,
		},
	}
}

type Serve struct {
	Database Database
	HTTP     HTTP
	Groups   Groups
}

type Migrate struct {
	Database Database
}

func Load[T any]() (T, error) {
	return env.ParseAs[T]()
}
