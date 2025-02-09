package containershandler

import (
	"app-pinger/backend/internal/usecase"
	queue "app-pinger/pkg/queue"
)

type ContainersHandler struct {
	containers usecase.ContainerRepo
	rabbitMQ   queue.RabbitMQ
}

func NewContainersHandler(c usecase.ContainerRepo, r queue.RabbitMQ) *ContainersHandler {
	return &ContainersHandler{
		containers: c,
		rabbitMQ:   r,
	}
}
