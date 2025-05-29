package config

import (
	"github.com/caarlos0/env/v10"
	"time"
	"user_svc/pkg/postgres"
)

type (
	Config struct {
		Postgres postgres.Config
		Server   Server
		Nats     Nats
		Redis    Redis
		Cache    Cache

		Version string `env:"VERSION"`
	}

	Server struct {
		GRPCServer GRPCServer
	}

	GRPCServer struct {
		Port                  int16         `env:"GRPC_PORT,notEmpty" envDefault:"8082"`
		MaxRecvMsgSizeMiB     int           `env:"GRPC_MAX_MESSAGE_SIZE_MIB" envDefault:"12"`
		MaxConnectionAge      time.Duration `env:"GRPC_MAX_CONNECTION_AGE" envDefault:"30s"`
		MaxConnectionAgeGrace time.Duration `env:"GRPC_MAX_CONNECTION_AGE_GRACE" envDefault:"10s"`
	}

	// Nats configuration for main application
	Nats struct {
		Hosts  []string `env:"NATS_HOSTS,notEmpty" envSeparator:"," envDefault:"localhost:4222"`
		NKey   string   `env:"NATS_NKEY,notEmpty"`
		IsTest bool     `env:"NATS_IS_TEST,notEmpty" envDefault:"true"`
	}

	// Redis configuration for main application
	Redis struct {
		Host         string        `env:"REDIS_HOSTS,notEmpty" envSeparator:"," envDefault:"localhost:6379"`
		Password     string        `env:"REDIS_PASSWORD" envDefault:""`
		TLSEnable    bool          `env:"REDIS_TLS_ENABLE" envDefault:"false"`
		DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" envDefault:"60s"`
		WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" envDefault:"60s"`
		ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" envDefault:"30s"`
	}

	Cache struct {
		ClientTTL time.Duration `env:"REDIS_CACHE_CLIENT_TTL" envDefault:"24h"`

		CMSVariableRefreshTime time.Duration `env:"CLIENT_REFRESH_TIME" envDefault:"1m"`
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
