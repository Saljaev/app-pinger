package service

import (
	"app-pinger/pkg/contracts"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-ping/ping"
	"log/slog"
	"net/http"
	"strings"
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
	GetIPs(list []string, whiteList bool) []string
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
func (p *PingerSvc) GetIPs(list []string, whiteList bool) []string {
	return p.Pinger.GetIPs(list, whiteList)
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
	Network      []string
}

// check for implementation
var _ Pinger = (*GoPinger)(nil)

func NewGoPingerService(cli *client.Client, log *slog.Logger, pCount int, pTimeout time.Duration, containerName string) *GoPinger {
	pinger := &GoPinger{
		cli:          cli,
		log:          *log,
		packetsCount: pCount,
		pingTimeout:  pTimeout,
	}

	pinger.searchNetworkByName(containerName)

	return pinger
}

// SearchNetworkByName добавляет в сервис все сети, которые указаны у него в docker-compose.yaml
func (p *GoPinger) searchNetworkByName(containerName string) {
	containers, err := p.getContainerList()
	if err != nil {
		p.log.Error("failed to get container list", slog.Any("error", err))
		return
	}

	var containerID string

	for _, container := range containers {
		if p.extractContainerName(container) == containerName {
			containerID = container.ID
			break
		}
	}

	nets, err := p.inspectContainer(containerID)
	if err != nil {
		p.log.Error("failed to inspect container", slog.String("ID", containerID), slog.Any("error", err))
		return
	}

	for net := range nets.NetworkSettings.Networks {
		p.Network = append(p.Network, net)
	}
}

// GetIPs возвращает список IP-адресов с учетом фильтра, а также его типом (белый/черный список)
func (p *GoPinger) GetIPs(list []string, whiteList bool) []string {
	p.log.Debug("starting get container list")

	containers, err := p.getContainerList()
	if err != nil {
		p.log.Error("failed to get container list", slog.Any("error", err))
		return nil
	}
	if len(containers) == 0 {
		p.log.Error("failed to get container list", slog.Any("error", "get 0 containers"))
		return nil
	}

	hasFilter := len(list) > 0

	p.log.Debug("successful get container list")

	listFilter := make(map[string]struct{}, len(list))
	for _, container := range list {
		listFilter[container] = struct{}{}
	}

	var ips []string
	for _, container := range containers {
		inspect, err := p.inspectContainer(container.ID)
		if err != nil {
			p.log.Error("failed to inspect container", slog.String("ID", container.ID), slog.Any("error", err))
			continue
		}

		containerIPs := p.filterNetworkIPs(inspect, p.Network)
		if len(containerIPs) == 0 {
			continue
		}

		containerName := p.extractContainerName(container)

		if p.shouldInclude(containerName, containerIPs, listFilter, hasFilter, whiteList) {
			ips = append(ips, containerIPs...)
		}
	}
	p.log.Debug("successful get containers", slog.Any("containers ips:", ips))
	return ips
}

// getContainerList получает список всех найденных контейнеров
func (p *GoPinger) getContainerList() ([]types.Container, error) {
	containers, err := p.cli.ContainerList(context.Background(), containertypes.ListOptions{})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// inspectContainer проверяет контейнер по айди и возвращает его параметы
func (p *GoPinger) inspectContainer(id string) (types.ContainerJSON, error) {
	inspect, err := p.cli.ContainerInspect(context.Background(), id)
	if err != nil {
		return types.ContainerJSON{}, err
	}

	return inspect, nil
}

// extractContainerName возвращает имя контейнера без префикса '/'
func (p *GoPinger) extractContainerName(c types.Container) string {
	if len(c.Names) == 0 {
		p.log.Error("unnamed container detected", slog.String("ID", c.ID))
		return ""
	}
	return strings.TrimPrefix(c.Names[0], "/")
}

// filterNetworkIPs возвращает список IP-адресов контейнера info, который находится в одной сети с networks
func (p *GoPinger) filterNetworkIPs(container types.ContainerJSON, networks []string) []string {
	var ips []string
	for netName, netSettings := range container.NetworkSettings.Networks {
		for _, network := range networks {
			if netName == network {
				ips = append(ips, netSettings.IPAddress)
			}
		}

	}
	return ips
}

// shouldInclude фильрация IP адресов ips в соответствии с фильтром filter и его типом whiteList
func (p *GoPinger) shouldInclude(name string, ips []string, filter map[string]struct{}, hasFilter, whitelist bool) bool {
	if !hasFilter {
		return true
	}

	for _, ip := range ips {
		_, matchIP := filter[ip]
		_, matchName := filter[name]

		if whitelist {
			if matchIP || matchName {
				return true
			}
		} else {
			if matchIP || matchName {
				return false
			}
		}
	}

	return !whitelist
}

// Ping пингует IP-адрес и возвращает данные о доступности
func (p *GoPinger) Ping(IP string) contracts.PingData {
	p.log.Debug("starting ping", slog.String("IP", IP))

	pinger, err := ping.NewPinger(IP)
	if err != nil {
		p.log.Error("failed to ping ", slog.String("IP", IP), slog.Any("error", err))
		return newPingData(IP, false, time.Now(), 0)
	}

	pinger.Count = p.packetsCount
	pinger.Timeout = p.pingTimeout

	pinger.Run()
	stats := pinger.Statistics()

	if stats.PacketsRecv > 0 {
		p.log.Debug("successful ping", slog.String("IP", IP), slog.Any("PacketsSend", pinger.Count),
			slog.Any("PacketsReceived", pinger.PacketsRecv))
		return newPingData(IP, true, time.Now(), stats.PacketLoss)
	}

	return newPingData(IP, false, time.Now(), stats.PacketLoss)
}

// SendRequest отправляет запрос на адрес url с информацией о пингах data
func (p *GoPinger) SendRequest(url string, data []contracts.PingData) error {
	req := contracts.ContainerAddReq{Containers: data}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to decode json :%w", err)
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {

	}

	return nil
}
