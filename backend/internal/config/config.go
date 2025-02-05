package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type Config struct {
	StoragePath string
	Addr        string        `env:"ADDRESS"`
	Timeout     time.Duration `env:"TIMEOUT"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT"`
	LogLevel    string        `env:"BACKEND_LOG_LEVEL"`
	User        string        `env:"POSTGRES_USER"`
	Password    string        `env:"POSTGRES_PASSWORD"`
	Host        string        `env:"PG_CONTAINER"`
	DB          string        `env:"DB"`
}

func ConfigLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, trying to load from environment variables")
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Println("failed to read .env file, default value have been used")
	}

	cfg.StoragePath = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.DB)

	return &cfg
}
