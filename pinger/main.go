package main

import (
	"app-pinger/pinger/config"
	"app-pinger/pinger/service"
	"app-pinger/pkg/contracts"
	"app-pinger/pkg/loger"
	queue "app-pinger/pkg/quque"
	_ "embed"
	"github.com/docker/docker/client"
	"log/slog"
	"strings"
	"sync"
	"time"
)

//go:embed list.txt
var filterList string

func main() {
	cfg := config.ConfigLoad()

	containerACL := strings.Split(filterList, "\n")
	whiteList := true

	var list []string
	if len(containerACL) > 0 {
		if containerACL[0] == "black" {
			whiteList = false
		}

		list = containerACL[1:]
	}

	log := loger.SetupLogger(cfg.LogLevel)

	log.Info("starting pinger-server")
	log.Debug("debug message are enabled")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error("failed to open API client", slog.Any("error", err))
	}

	defer cli.Close()

	rabbitMQ, err := queue.NewConnection(cfg.RabbitMQPath, cfg.RabbitMQ.Queue)
	if err != nil {
		log.Error("failed to create rabbitMQ connection", slog.Any("error", err))
	}
	defer rabbitMQ.Close()

	pinger := service.NewPingerService(service.NewGoPingerService(cli, log, cfg.PacketsCount, cfg.PingTimeout, cfg.ServiceName, *rabbitMQ))

	log.Info("pinger-server started")
	log.Debug("service settings", slog.Any("service-timeout", cfg.SvcTimeout),
		slog.Any("ping-packets", cfg.PacketsCount), slog.Any("ping-timeout", cfg.PingTimeout),
		slog.Any("network", cfg.Network))

	reach := make(map[string]contracts.PingData)

	ticker := time.NewTicker(cfg.SvcTimeout)
	//TODO: change for ticker
	for range ticker.C {
		netIPs := pinger.GetIPs(list, whiteList)

		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}

		for net, ips := range netIPs {
			for _, ip := range ips {
				wg.Add(1)

				go func(net, ip string) {
					defer wg.Done()
					data := pinger.Ping(net, ip)
					mutex.Lock()
					reach[ip] = data

					mutex.Unlock()
				}(net, ip)
			}
			wg.Wait()
		}
		wg.Wait()
		pingArr := make([]contracts.PingData, 0, len(reach))
		for _, v := range reach {
			pingArr = append(pingArr, v)
		}

		err = pinger.SendRequest(pingArr)
		if err != nil {
			log.Error("failed to send request", slog.Any("error", err))
		}
	}
}
