package testcontainer

import (
	"context"
)

// PostgresConfig - postgres container config.
type PostgresConfig struct {
	BaseContainerConfig
}

// NewPostgresContainer - create postgres container, but do not start it yet.
func NewPostgresContainer(
	ctx context.Context,
	cfg PostgresConfig,
	dockerNetwork *DockerNetwork,
) (Container, error) {
	return NewGenericContainer(ctx, cfg.BaseContainerConfig, dockerNetwork)
}
