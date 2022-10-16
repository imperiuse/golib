package testcontainer

import (
	"context"
)

// ZookeeperConfig - zookeeper container config.
type ZookeeperConfig struct {
	BaseContainerConfig
}

// NewZookeeperContainer - create zookeeper container, but do not start it yet.
func NewZookeeperContainer(
	ctx context.Context,
	cfg ZookeeperConfig,
	dockerNetwork *DockerNetwork,
) (Container, error) {
	// creates the zookeeper container, but do not start it yet
	return NewGenericContainer(ctx, cfg.BaseContainerConfig, dockerNetwork)
}
