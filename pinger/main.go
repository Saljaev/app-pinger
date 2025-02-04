package main

import (
	"app-pinger/pinger/service"
	"app-pinger/pkg/loger"
	"github.com/docker/docker/client"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"log/slog"
	"sync"
	"time"
)

type ConfigPingerSvc struct {
	LogLevel     string        `env:"PINGER_LOG_LEVEL" env-default:"info"`
	PacketsCount int           `env:"PINGER_PACKETS_COUNT" env-default:"4"`
	PingTimeout  time.Duration `env:"PINGER_PING_TIMEOUT" env-default:"5s"`
	SvcTimeout   time.Duration `env:"PINGER_SVC_PING_TIMEOUT" env-default:"10s"`
}

type Data struct {
	isReachable bool
	packetLoss  float64
}

func main() {
	var cfg ConfigPingerSvc

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Println("failed to read .env file, default value have been used")
	}

	log := loger.SetupLogger(cfg.LogLevel)

	log.Info("starting Pinger-Service")
	log.Debug("debug message are enabled")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error("failed to open API client", slog.Any("error", err))
	}

	defer cli.Close()

	pinger := service.NewPingerService(service.NewGoPingerService(cli, log, cfg.PacketsCount, cfg.PingTimeout))

	log.Info("pinger-service started")
	log.Debug("service settings", slog.Any("service-timeout", cfg.SvcTimeout),
		slog.Any("ping-packets", cfg.PacketsCount), slog.Any("ping-timeout", cfg.PingTimeout))

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
