package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
)

type DataBase struct {
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Host     string `env:"PG_CONTAINER"`
	DB       string `env:"DB"`
}

type RabbitMQ struct {
	User     string `env:"RABBITMQ_USER"`
	Password string `env:"RABBITMQ_PASS"`
	Host     string `env:"RABBITMQ_HOST"`
	Queue    string `env:"RABBITMQ_QUEUE"`
}

func ConfigLoad(cfg interface{}) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, trying to load from environment variables")
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Println("failed to read .env file")
	}
}

func (d *DataBase) NewDBPath() string {
	return fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", d.User, d.Password,
		d.Host, d.DB)
}

func (r *RabbitMQ) NewRabbitMQPath() string {
	return fmt.Sprintf("amqp://%s:%s@%s:5672/", r.User, r.Password, r.Host)
}
