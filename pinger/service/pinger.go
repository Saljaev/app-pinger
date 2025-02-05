package service

import (
	"app-pinger/pkg/contracts"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-ping/ping"
	"log/slog"
	"net/http"
	"time"
)

func newPingData(IP string, isReachable bool, LastPing time.Time, PacketLost float64) contracts.PingData {
	return contracts.PingData{
		IPAddress:   IP,
		IsReachable: isReachable,
		LastPing:    LastPing.Format(time.DateTime),
		PackerLost:  PacketLost,
	}
}

// Pinger интерфейс, который определяет логику сервиса
type Pinger interface {
	GetIPs() []string
	Ping(IP string) contracts.PingData
	SendRequest(url string, data []contracts.PingData) error
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
func (p *PingerSvc) Ping(IP string) contracts.PingData {
	return p.Pinger.Ping(IP)
}

// SendRequest отправляет запрос к backend-svc с данными ping всех контейнеров
func (p *PingerSvc) SendRequest(url string, data []contracts.PingData) error {
	return p.Pinger.SendRequest(url, data)
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
	p.log.Debug("starting get container list")
	containers, err := p.cli.ContainerList(context.Background(), containertypes.ListOptions{})
	if err != nil {
		p.log.Error("failed to get container list", slog.Any("error", err))
		return nil
	}

	p.log.Debug("successful get container list")

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
	p.log.Debug("successful get IP-address", slog.Any("IP-address count", len(ips)))
	return ips
}

// TODO: add white/black list for ip
func (p *GoPinger) Ping(IP string) contracts.PingData {
	p.log.Debug("starting ping", slog.String("IP", IP))

	pinger, err := ping.NewPinger(IP)
	if err != nil {
		p.log.Error("failed to ping ", slog.String("IP", IP), slog.Any("error", err))
		return newPingData(IP, false, time.Now(), 0)
	}

	pinger.Count = p.packetsCount
	pinger.Timeout = p.pingTimeout

	p.log.Debug("ping settings", slog.Any("Packets count", pinger.Count),
		slog.Any("Timeout", pinger.Timeout.Seconds()))

	pinger.Run()
	stats := pinger.Statistics()

	if stats.PacketsRecv > 0 {
		p.log.Debug("successful ping", slog.String("IP", IP), slog.Any("PacketsSend", pinger.Count),
			slog.Any("PacketsReceived", pinger.PacketsRecv))
		return newPingData(IP, true, time.Now(), stats.PacketLoss)
	}

	return newPingData(IP, false, time.Now(), stats.PacketLoss)
}

func (p *GoPinger) SendRequest(u string, data []contracts.PingData) error {
	req := contracts.ContainerAddReq{Containers: data}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to decode json :%w", err)
	}

	_, err = http.Post(u, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {

	}

	return nil
}
