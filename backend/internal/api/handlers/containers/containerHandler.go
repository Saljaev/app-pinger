package containershandler

import (
	"app-pinger/backend/internal/usecase"
)

type ContainersHandler struct {
	containers usecase.ContainerRepo
}

func NewContainersHandler(c usecase.ContainerRepo) *ContainersHandler {
	return &ContainersHandler{
		containers: c,
	}
}
