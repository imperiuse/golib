package testcontainer

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// RedisConfig - redis container config.
type RedisConfig struct {
	BaseContainerConfig
}

// NewRedisContainer - create redis container, but do not start it yet.
func NewRedisContainer(
	ctx context.Context,
	cfg NginxConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) (Container, error) {
	return NewGenericContainer(ctx, cfg.BaseContainerConfig, dockerNetwork)
}
