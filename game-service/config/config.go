package config

import (
	"game_svc/pkg/redis"
	"github.com/caarlos0/env/v10"
	"time"
)

type (
	Config struct {
		Redis      redis.Config
		Server     ServerConfig
		Nats       Nats
		JWTManager JWTManager
		GRPC       GRPC
		Version    string `env:"VERSION"`
	}

	ServerConfig struct { // Renamed for clarity
		WebSocketPort      string        `env:"WEBSOCKET_PORT,notEmpty" envDefault:"8081"`
		WebSocketPath      string        `env:"WEBSOCKET_PATH,notEmpty" envDefault:"/ws"`
		ReadTimeoutSec     time.Duration `env:"WEBSOCKET_READ_TIMEOUT_SEC" envDefault:"60s"`  // e.g. for upgrader or http server
		WriteTimeoutSec    time.Duration `env:"WEBSOCKET_WRITE_TIMEOUT_SEC" envDefault:"10s"` // e.g. for upgrader or http server
		IdleTimeoutSec     time.Duration `env:"WEBSOCKET_IDLE_TIMEOUT_SEC" envDefault:"120s"` // e.g. for http server
		ShutdownTimeoutSec time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT_SEC" envDefault:"15s"`
		// AllowedOrigins  []string `env:"WEBSOCKET_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"` // For CheckOrigin
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
		GameResultSubject string `env:"NATS_GAME_RESULT_SUBJECT,notEmpty"`
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

	JWTManager struct {
		SecretKey string `env:"JWT_MANAGER_SECRET_KEY,notEmpty"`
	}
	GRPC struct {
		GRPCClient GRPCClient
	}

	GRPCClient struct {
		UserServiceURL string `env:"GRPC_USER_SERVICE_URL,required"`
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
