package containershandler

import (
	"app-pinger/backend/internal/entity"
	"app-pinger/pkg/contracts"
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

type ContainerAddResp struct {
	Text string `json:"msg"`
}

func (c *ContainersHandler) ProcessQueue(log *slog.Logger) {
	msgs, err := c.rabbitMQ.Consume()
	if err != nil {
		log.Error("failed get messages from rabbitmq", err)
		return
	}

	for msg := range msgs {
		var req contracts.ContainerAddReq

		if err := json.Unmarshal(msg.Body, &req); err != nil {
			log.Error("failed to decode RabbitMQ message", err)
			continue
		}

		if !req.IsValid() {
			log.Error("failed decode json", slog.Any("error", "invalid request"))
		}

		log.Debug("received request", "request", req)

		for _, r := range req.Containers {
			lastPing, err := time.Parse(time.DateTime, r.LastPing)
			if err != nil {
				log.Error("failed encode containers", err)
				return
			}

			container := entity.Container{
				IP:          r.IPAddress,
				IsReachable: r.IsReachable,
				LastPing:    lastPing,
				PacketLost:  r.PackerLost,
			}

			IP, err := c.containers.Add(context.Background(), container)
			if err != nil || IP != r.IPAddress {
				log.Error("failed to add container", err)
				return
			}
		}
	}
}
