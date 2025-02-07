package config

import (
	"app-pinger/pkg/config"
	"time"
)

// TODO: add MAKEFILE to generate default .env
type Config struct {
	LogLevel     string        `env:"PINGER_LOG_LEVEL"`
	PacketsCount int           `env:"PINGER_PACKETS_COUNT"`
	PingTimeout  time.Duration `env:"PINGER_PING_TIMEOUT"`
	SvcTimeout   time.Duration `env:"PINGER_SVC_PING_TIMEOUT" `
	BackendName  string        `env:"BACKEND_HOST"`
	ServiceName  string        `env:"PINGER_HOST"`
	BackendPort  string        `env:"BACKEND_PORT"`
	Network      string        `env:"PINGER_NETWORK"`
	RabbitMQPath string
	RabbitMQ     config.RabbitMQ
}

func ConfigLoad() *Config {
	var cfg Config

	config.ConfigLoad(&cfg)

	cfg.RabbitMQPath = cfg.RabbitMQ.NewRabbitMQPath()

	return &cfg
}
