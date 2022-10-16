package testcontainer

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// NginxConfig - nginx container config.
type NginxConfig struct {
	BaseContainerConfig
}

// NewNginxContainer - create nginx container, but do not start it yet.
func NewNginxContainer(
	ctx context.Context,
	cfg NginxConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) (Container, error) {
	return NewGenericContainer(ctx, cfg.BaseContainerConfig, dockerNetwork)
}
