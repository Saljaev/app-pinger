package service

import (
	"app-pinger/pkg/contracts"
	queue "app-pinger/pkg/queue"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/go-ping/ping"
	"log/slog"
	"strings"
	"sync"
	"time"
)

func newPingData(IP string, isReachable bool, LastPing time.Time) contracts.PingData {
	return contracts.PingData{
		IPAddress:   IP,
		IsReachable: isReachable,
		LastPing:    LastPing.Format(time.DateTime),
	}
}

// Pinger интерфейс, который определяет логику сервиса
type Pinger interface {
	GetIPs(list []string, whiteList bool) map[string][]string
	Ping(net, IP string) contracts.PingData
	SendRequest(data []contracts.PingData) error
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
func (p *PingerSvc) GetIPs(list []string, whiteList bool) map[string][]string {
	return p.Pinger.GetIPs(list, whiteList)
}

// Ping пингует IP адрес, возвращает данные доступности
func (p *PingerSvc) Ping(net, IP string) contracts.PingData {
	return p.Pinger.Ping(net, IP)
}

// SendRequest отправляет запрос к backend-svc через RabbitMQ с данными ping всех контейнеров
func (p *PingerSvc) SendRequest(data []contracts.PingData) error {
	return p.Pinger.SendRequest(data)
}

// GoPinger реализация PingerSvc, основанная на Docker SDK и go-ping
type GoPinger struct {
	cli          *client.Client
	log          slog.Logger
	packetsCount int
	pingTimeout  time.Duration
	rabbitMQ     queue.RabbitMQConnection
	id           string
	name         string
	net          map[string]struct{}
	mu           sync.Mutex
}

// check for implementation
var _ Pinger = (*GoPinger)(nil)

func NewGoPingerService(
	c *client.Client,
	l *slog.Logger,
	pC int,
	pT time.Duration,
	n string,
	r queue.RabbitMQConnection,
) *GoPinger {
	pinger := &GoPinger{
		cli:          c,
		log:          *l,
		packetsCount: pC,
		pingTimeout:  pT,
		rabbitMQ:     r,
		name:         n,
		net:          map[string]struct{}{},
		mu:           sync.Mutex{},
	}

	pinger.searchOwnIDAndNetwork()

	return pinger
}

// searchOwnIDAndNetwork находит и устанавливает id и сеть контейнера net по своему имени name
func (p *GoPinger) searchOwnIDAndNetwork() {
	containers, err := p.getContainerList()
	if err != nil {
		p.log.Error("failed to get container list", slog.Any("error", err))
		return
	}

	for _, container := range containers {
		if p.extractContainerName(container) == p.name {
			p.id = container.ID
			for _, net := range container.NetworkSettings.Networks {
				p.net[net.NetworkID] = struct{}{}
			}
			break
		}
	}
}

// GetIPs возвращает мапу сеть-IP-адреса с учетом фильтра, а также его типом (белый/черный список)
func (p *GoPinger) GetIPs(list []string, whiteList bool) map[string][]string {
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

	ips := map[string][]string{}
	for _, container := range containers {
		inspect, err := p.inspectContainer(container.ID)
		if err != nil {
			p.log.Error("failed to inspect container", slog.String("ID", container.ID), slog.Any("error", err))
			continue
		}

		containerIPs := p.filterNetworkIPs(inspect)
		if len(containerIPs) == 0 {
			continue
		}

		containerName := p.extractContainerName(container)

		for netName, ipList := range containerIPs {
			if p.shouldInclude(containerName, ipList, listFilter, hasFilter, whiteList) {
				ips[netName] = append(ips[netName], ipList...)
			}
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

// filterNetworkIPs возвращает список всех IP-адресов всех контейнеров
func (p *GoPinger) filterNetworkIPs(container types.ContainerJSON) map[string][]string {
	ips := map[string][]string{}
	for netName, netSettings := range container.NetworkSettings.Networks {
		ips[netSettings.NetworkID] = append(ips[netName], netSettings.IPAddress)
	}
	p.log.Debug("get ips", slog.Any("ips", ips))

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

// Ping пингует IP-адрес и возвращает данные о доступности контейнера в указанной сети
func (p *GoPinger) Ping(net, IP string) contracts.PingData {
	p.log.Debug("starting ping", slog.String("network", net), slog.String("IP", IP))
	err := p.connectToNetwork(net)
	if err != nil {
		p.log.Error("failed to switch network", slog.String("network", net), slog.Any("error", err))
		return newPingData(IP, false, time.Now())
	}

	pinger, err := ping.NewPinger(IP)
	if err != nil {
		p.log.Error("failed to ping ", slog.String("IP", IP), slog.Any("error", err))
		return newPingData(IP, false, time.Now())
	}

	pinger.Count = p.packetsCount
	pinger.Timeout = p.pingTimeout

	pinger.Run()
	stats := pinger.Statistics()

	if stats.PacketsRecv > 0 {
		p.log.Debug("successful ping", slog.String("IP", IP), slog.Any("PacketsSend", pinger.Count),
			slog.Any("PacketsReceived", pinger.PacketsRecv))
		return newPingData(IP, true, time.Now())
	}

	return newPingData(IP, false, time.Now())
}

// connectToNetwork подключает pinger к сети указанной сети
func (p *GoPinger) connectToNetwork(net string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.net[net]; ok {
		return nil
	}
	err := p.cli.NetworkConnect(context.Background(), net, p.id, &network.EndpointSettings{})
	if err != nil {
		return fmt.Errorf("failed to connect network: %w", err)
	}
	p.net[net] = struct{}{}
	return nil
}

// SendRequest отправляет запрос на адрес rabbitmq с информацией о пингах data
func (p *GoPinger) SendRequest(data []contracts.PingData) error {
	req := contracts.ContainerAddReq{Containers: data}

	err := p.rabbitMQ.Publish(req)
	if err != nil {
		return fmt.Errorf("failed to publish data: %w", err)
	}

	return nil
}
