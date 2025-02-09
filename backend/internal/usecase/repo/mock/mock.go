package storagemock

import (
	"app-pinger/backend/internal/entity"
	"app-pinger/backend/internal/usecase"
	"context"
)

type MockRepo struct {
	container entity.Container
}

// check for implementation
var _ usecase.ContainerRepo = (*MockRepo)(nil)

func NewMockRepo(c entity.Container) *MockRepo {
	return &MockRepo{container: c}
}

func (m *MockRepo) Add(ctx context.Context, container entity.Container) (string, error) {
	return container.IP, nil
}

func (m MockRepo) GetAll(ctx context.Context) ([]entity.Container, error) {
	return []entity.Container{m.container}, nil
}
