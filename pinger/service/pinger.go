package service

import (
	"context"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-ping/ping"
	"log/slog"
	"time"
)

// Pinger интерфейс, который определяет логику сервиса
type Pinger interface {
	GetIPs() []string
	Ping(IP string) (bool, float64)
}

// PingerSvc сервис, который выполняет бизнес логику сервиса
type PingerSvc struct {
	Pinger Pinger
}

func NewPingerService(pinger Pinger) *PingerSvc {
	return &PingerSvc{
		Pinger: pinger,
	}
}

// GetIPs получает все доступные IP-адреса контейнеров
func (p *PingerSvc) GetIPs() []string {
	return p.Pinger.GetIPs()
}

// Ping пингует IP адрес, возвращает доступность адрес и процент потери пакетов
func (p *PingerSvc) Ping(IP string) (bool, float64) {
	return p.Pinger.Ping(IP)
}

// GoPinger реализация PingerSvc, основанная на Docker SDK и go-ping
type GoPinger struct {
	cli          *client.Client
	log          slog.Logger
	packetsCount int
	pingTimeout  time.Duration
}

// check for implementation
var _ Pinger = (*GoPinger)(nil)

func NewGoPingerService(cli *client.Client, log *slog.Logger, pCount int, pTimeout time.Duration) *GoPinger {
	return &GoPinger{
		cli:          cli,
		log:          *log,
		packetsCount: pCount,
		pingTimeout:  pTimeout,
	}
}

func (p *GoPinger) GetIPs() []string {
	p.log.Debug("Starting get container list")
	containers, err := p.cli.ContainerList(context.Background(), containertypes.ListOptions{})
	if err != nil {
		p.log.Error("failed to get container list", slog.Any("error", err))
		return nil
	}

	p.log.Debug("Successful get container list")

	var ips []string
	for _, container := range containers {
		inspect, err := p.cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			p.log.Error("failed to inspect container", slog.String("ID", container.ID), slog.Any("error", err))
			continue
		}

		for _, netSettings := range inspect.NetworkSettings.Networks {
			ips = append(ips, netSettings.IPAddress)
		}
	}
	p.log.Debug("Successful get IP-address", slog.Any("IP-address count", len(ips)))
	return ips
}

func (p *GoPinger) Ping(IP string) (bool, float64) {
	p.log.Debug("Starting ping", slog.String("IP", IP))
	pinger, err := ping.NewPinger(IP)
	if err != nil {
		p.log.Error("failed to ping ", slog.String("IP", IP), slog.Any("error", err))
		return false, 0
	}

	pinger.Count = p.packetsCount
	pinger.Timeout = p.pingTimeout

	p.log.Debug("Ping settings", slog.Any("Packets count", pinger.Count),
		slog.Any("Timeout", pinger.Timeout.Seconds()))

	pinger.Run()
	stats := pinger.Statistics()

	if stats.PacketsRecv > 0 {
		p.log.Info("Successful ping", slog.String("IP", IP), slog.Any("PacketsSend", pinger.Count),
			slog.Any("PacketsReceived", pinger.PacketsRecv))
		return true, stats.PacketLoss
	}

	return false, 0
}
