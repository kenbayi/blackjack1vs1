package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		Server  Server
		GRPC    GRPC
		Version string `env:"VERSION"`
	}

	Server struct {
		HTTPServer HTTPServer
		JWTManager JWTManager
	}

	HTTPServer struct {
		Port           int           `env:"HTTP_PORT,required" envDefault:"8079"`
		ReadTimeout    time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
		WriteTimeout   time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
		IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
		MaxHeaderBytes int           `env:"HTTP_MAX_HEADER_BYTES" envDefault:"1048576"` // 1 MB
		TrustedProxies []string      `env:"HTTP_TRUSTED_PROXIES" envSeparator:","`
		Mode           string        `env:"GIN_MODE" envDefault:"release"` // Can be: release, debug, test
	}

	GRPC struct {
		GRPCClient GRPCClient
	}

	GRPCClient struct {
		AuthServiceURL       string `env:"GRPC_AUTH_SERVICE_URL,required" envDefault:"0.0.0.0:8080"`
		StatisticsServiceURL string `env:"GRPC_STATISTICS_SERVICE_URL,required" envDefault:"0.0.0.0:8083"`
		UserServiceURL       string `env:"GRPC_USER_SERVICE_URL,required" envDefault:"0.0.0.0:8082"`
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
