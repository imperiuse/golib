package testcontainer

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// ZookeeperConfig - zookeeper container config.
type ZookeeperConfig struct {
	BaseContainerConfig
}

// NewZookeeperContainer - create zookeeper container, but do not start it yet.
func NewZookeeperContainer(
	ctx context.Context,
	cfg ZookeeperConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) (testcontainers.Container, error) {
	cfg.ExposedPorts = []string{cfg.Port}
	cfg.Envs = map[string]string{"ZOOKEEPER_CLIENT_PORT": cfg.Port, "ZOOKEEPER_TICK_TIME": "2000"}

	// creates the zookeeper container, but do not start it yet
	return testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: GetBaseContainerRequest(cfg.BaseContainerConfig, dockerNetwork),
		},
	)
}
