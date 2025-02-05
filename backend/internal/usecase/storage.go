package usecase

import (
	"app-pinger/backend/internal/entity"
	"context"
	"fmt"
)

type ContainerRepo interface {
	Add(ctx context.Context, c entity.Container) (string, error)
	GetAll(ctx context.Context) ([]entity.Container, error)
}

type BackendService struct {
	repo ContainerRepo
}

func NewBackendService(repo ContainerRepo) *BackendService {
	return &BackendService{repo: repo}
}

func (b *BackendService) Add(ctx context.Context, c entity.Container) (string, error) {
	const op = "BackendService - Add"

	IP, err := b.repo.Add(ctx, c)
	if err != nil {
		return "", fmt.Errorf("%s - b.repo.Add: %w", op, err)
	}

	return IP, nil
}

func (b *BackendService) GetAll(ctx context.Context) ([]entity.Container, error) {
	const op = "BackendService - GetAll"

	containers, err := b.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s - b.repo.GetAll: %w", op, err)
	}

	return containers, nil
}
