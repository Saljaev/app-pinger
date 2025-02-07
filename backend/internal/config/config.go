package config

import (
	"app-pinger/pkg/config"
	"fmt"
	"time"
)

type Config struct {
	RabbitMQPath string
	StoragePath  string
	Addr         string        `env:"BACKEND_ADDRESS"`
	Port         string        `env:"BACKEND_PORT"`
	Timeout      time.Duration `env:"TIMEOUT"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT"`
	LogLevel     string        `env:"BACKEND_LOG_LEVEL"`
	DB           config.DataBase
	RabbitMQ     config.RabbitMQ
}

func ConfigLoad() *Config {
	var cfg Config

	config.ConfigLoad(&cfg)

	cfg.StoragePath = cfg.DB.NewDBPath()
	cfg.RabbitMQPath = cfg.RabbitMQ.NewRabbitMQPath()
	cfg.Addr = fmt.Sprintf("%s:%s", cfg.Addr, cfg.Port)

	return &cfg
}
