package config

import (
	"auth_svc/pkg/postgres"
	"github.com/caarlos0/env/v10"
	"time"
)

type (
	Config struct {
		Postgres   postgres.Config
		Server     Server
		Nats       Nats
		JWTManager JWTManager
		Redis      Redis
		EmailRedis EmailRedis
		Version    string `env:"VERSION"`
	}

	Server struct {
		GRPCServer GRPCServer
	}

	GRPCServer struct {
		Port                  int16         `env:"GRPC_PORT,notEmpty" envDefault:"8080"`
		MaxRecvMsgSizeMiB     int           `env:"GRPC_MAX_MESSAGE_SIZE_MIB" envDefault:"12"`
		MaxConnectionAge      time.Duration `env:"GRPC_MAX_CONNECTION_AGE" envDefault:"30s"`
		MaxConnectionAgeGrace time.Duration `env:"GRPC_MAX_CONNECTION_AGE_GRACE" envDefault:"10s"`
	}

	// Nats configuration for main application
	Nats struct {
		Hosts        []string `env:"NATS_HOSTS,notEmpty" envSeparator:"," envDefault:"localhost:4222"`
		NKey         string   `env:"NATS_NKEY,notEmpty"`
		IsTest       bool     `env:"NATS_IS_TEST,notEmpty" envDefault:"true"`
		NatsSubjects NatsSubjects
	}

	// NatsSubjects for main application
	NatsSubjects struct {
		UserCreatedSubject string `env:"NATS_USER_CREATED_SUBJECT,notEmpty"`
		UserUpdatedSubject string `env:"NATS_USER_UPDATED_SUBJECT,notEmpty"`
		UserDeletedSubject string `env:"NATS_USER_DELETED_SUBJECT,notEmpty"`
		EmailChangeSubject string `env:"NATS_EMAIL_CHANGE_SUBJECT,notEmpty"`
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

	EmailRedis struct {
		ClientTTL              time.Duration `env:"REDIS_CACHE_CLIENT_TTL" envDefault:"24h"`
		CMSVariableRefreshTime time.Duration `env:"CLIENT_REFRESH_TIME" envDefault:"1m"`
	}

	JWTManager struct {
		SecretKey string `env:"JWT_MANAGER_SECRET_KEY,notEmpty"`
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
