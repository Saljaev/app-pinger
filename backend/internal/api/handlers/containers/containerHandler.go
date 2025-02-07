package containershandler

import (
	"app-pinger/backend/internal/usecase"
	queue "app-pinger/pkg/quque"
)

type ContainersHandler struct {
	containers usecase.ContainerRepo
	rabbitMQ   queue.RabbitMQConnection
}

func NewContainersHandler(c usecase.ContainerRepo, r queue.RabbitMQConnection) *ContainersHandler {
	return &ContainersHandler{
		containers: c,
		rabbitMQ:   r,
	}
}
