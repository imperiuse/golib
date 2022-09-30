package testcontainer

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NginxConfig - nginx container config.
type NginxConfig struct {
	BaseContainerConfig
}

type nginxContainer struct {
	testcontainers.Container
	URI string
}

// NewNginxContainer - create nginx container, but do not start it yet.
func NewNginxContainer(
	ctx context.Context,
	cfg NginxConfig,
	dockerNetwork *testcontainers.DockerNetwork,
	runContainer bool,
) (*nginxContainer, error) {
	cfg.ExposedPorts = []string{cfg.Port + "/tcp"}

	cr := GetBaseContainerRequest(cfg.BaseContainerConfig, dockerNetwork)

	cr.WaitingFor = wait.ForHTTP("/")

	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: cr,
			Started:          runContainer,
		},
	)
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(cfg.Port))
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return &nginxContainer{
		Container: container,
		URI:       uri,
	}, nil
}
