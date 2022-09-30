package testcontainer

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisConfig - redis container config.
type RedisConfig struct {
	BaseContainerConfig
}

type redisContainer struct {
	testcontainers.Container
	URI string
}

// NewRedisContainer - create redis container, but do not start it yet.
func NewRedisContainer(
	ctx context.Context,
	cfg NginxConfig,
	dockerNetwork *testcontainers.DockerNetwork,
	runContainer bool,
) (testcontainers.Container, error) {
	cfg.ExposedPorts = []string{cfg.Port + "/tcp"}

	cr := GetBaseContainerRequest(cfg.BaseContainerConfig, dockerNetwork)

	cr.WaitingFor = wait.ForLog("* Ready to accept connections")

	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: cr,
			Started:          runContainer,
		},
	)

	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(cfg.Port))
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("redis://%s:%s", hostIP, mappedPort.Port())

	return &redisContainer{Container: container, URI: uri}, nil
}
