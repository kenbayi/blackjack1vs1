package config

import (
	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		MailerKey string `env:"MAILER_API_KEY,notEmpty"`
		Nats      Nats

		Version string `env:"VERSION"`
	}

	// Nats configuration for main application
	Nats struct {
		Hosts  []string `env:"NATS_HOSTS,notEmpty" envSeparator:"," envDefault:"localhost:4222"`
		NKey   string   `env:"NATS_NKEY,notEmpty"`
		IsTest bool     `env:"NATS_IS_TEST,notEmpty" envDefault:"true"`
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	return &cfg, err
}
