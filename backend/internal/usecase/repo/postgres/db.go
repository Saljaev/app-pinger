package postgres

import (
	"app-pinger/backend/internal/entity"
	"app-pinger/backend/internal/usecase"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ContainerRepo struct {
	*sql.DB
}

// check for implementation
var _ usecase.ContainerRepo = (*ContainerRepo)(nil)

func NewContainerRepo(db *sql.DB) *ContainerRepo {
	return &ContainerRepo{db}
}

func (c *ContainerRepo) Add(ctx context.Context, container entity.Container) (string, error) {
	const op = "ContainerRepo - Add"

	query := "INSERT INTO containers(ip_address, is_reachable, last_ping) " +
		"VALUES($1, $2, $3) " +
		"ON CONFLICT(ip_address) " +
		"DO UPDATE SET " +
		"is_reachable = EXCLUDED.is_reachable, " +
		"last_ping = EXCLUDED.last_ping " +
		"WHERE containers.last_ping < EXCLUDED.last_ping " +
		"RETURNING ip_address"

	var containerID string

	err := c.QueryRowContext(ctx, query, container.IP, container.IsReachable, container.LastPing).Scan(&containerID)
	if errors.Is(err, sql.ErrNoRows) {
		return container.IP, nil
	}
	if err != nil {
		return "", fmt.Errorf("%s - c.QueryRowContext: %w", op, err)
	}

	return containerID, nil

}

func (c ContainerRepo) GetAll(ctx context.Context) ([]entity.Container, error) {
	const op = "ContainerRepo - GetAll"

	query := "SELECT ip_address, is_reachable, last_ping FROM containers"

	rows, err := c.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s - c.QueryContext: %w", op, err)
	}

	defer rows.Close()

	containers := []entity.Container{}

	for rows.Next() {
		var container entity.Container

		rows.Scan(&container.IP, &container.IsReachable, &container.LastPing)

		containers = append(containers, container)
	}

	return containers, nil
}
