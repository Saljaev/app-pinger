package main

import (
	"app-pinger/pinger/service"
	"app-pinger/pkg/contracts"
	"app-pinger/pkg/loger"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// TODO: add MAKEFILE to generate default .env
type ConfigPingerSvc struct {
	LogLevel     string        `env:"PINGER_LOG_LEVEL"`
	PacketsCount int           `env:"PINGER_PACKETS_COUNT"`
	PingTimeout  time.Duration `env:"PINGER_PING_TIMEOUT"`
	SvcTimeout   time.Duration `env:"PINGER_SVC_PING_TIMEOUT" `
	BackendName  string        `env:"BACKEND_NAME"`
	BackendIP    string        `env:"ADDRESS"`
}

func main() {
	var cfg ConfigPingerSvc

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Println("failed to read .env file, default value have been used")
	}

	cfg.BackendIP = strings.Split(cfg.BackendIP, ":")[1]

	log := loger.SetupLogger(cfg.LogLevel)

	log.Info("starting pinger-server")
	log.Debug("debug message are enabled")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error("failed to open API client", slog.Any("error", err))
	}

	defer cli.Close()

	pinger := service.NewPingerService(service.NewGoPingerService(cli, log, cfg.PacketsCount, cfg.PingTimeout))

	log.Info("pinger-server started")
	log.Debug("service settings", slog.Any("service-timeout", cfg.SvcTimeout),
		slog.Any("ping-packets", cfg.PacketsCount), slog.Any("ping-timeout", cfg.PingTimeout))

	for {
		//TODO: return name of container
		ips := pinger.GetIPs()

		//TODO: cache for ip containers
		reach := []contracts.PingData{}

		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}

		for _, ip := range ips {
			wg.Add(1)

			go func(ip string) {
				defer wg.Done()
				data := pinger.Ping(ip)
				mutex.Lock()
				reach = append(reach, data)
				mutex.Unlock()
			}(ip)
		}
		wg.Wait()

		err = pinger.SendRequest(fmt.Sprintf("http://%s:%s/container/add", cfg.BackendName, cfg.BackendIP), reach)
		if err != nil {
			log.Error("failed to send request", slog.Any("error", err))
		}

		time.Sleep(cfg.SvcTimeout)
	}
}
