package main

import (
	"app-pinger/pinger/service"
	"github.com/docker/docker/client"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
)

type ConfigPingerSvc struct {
	LogLevel     string        `env:"log_level" env-default:"info"`
	PacketsCount int           `env:"packets_count" env-default:"4"`
	PingTimeout  time.Duration `env:"ping_timeout" env-default:"5s"`
	SvcTimeout   time.Duration `env:"svc_ping_timeout" env-default:"10s"`
}

type Data struct {
	isReachable bool
	packetLoss  float64
}

func main() {
	var cfg ConfigPingerSvc

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Println("Failed to read .env file, default value have been used")
	}

	log := setupLogger(cfg.LogLevel)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error("Failed to open API client", slog.Any("error", err))
	}

	defer cli.Close()

	pinger := service.NewPingerService(service.NewGoPingerService(cli, log, cfg.PacketsCount, cfg.PingTimeout))

	for {
		ips := pinger.GetIPs()

		reach := make(map[string]Data)

		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}

		for _, ip := range ips {
			wg.Add(1)

			go func(ip string) {
				defer wg.Done()
				reachable, lossPacket := pinger.Ping(ip)
				mutex.Lock()
				reach[ip] = Data{
					isReachable: reachable,
					packetLoss:  lossPacket,
				}
				mutex.Unlock()
			}(ip)
		}

		wg.Wait()

		time.Sleep(cfg.SvcTimeout)

	}
}

func setupLogger(level string) *slog.Logger {
	var log *slog.Logger

	switch level {
	case LevelInfo:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case LevelDebug:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return log
}
