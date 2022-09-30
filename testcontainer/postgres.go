package testcontainer

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresConfig - postgres container config.
type PostgresConfig struct {
	BaseContainerConfig
}

type postgresContainer struct {
	testcontainers.Container
	URI string
}

// NewPostgresContainer - create nginx container, but do not start it yet.
func NewPostgresContainer(
	ctx context.Context,
	cfg PostgresConfig,
	dockerNetwork *testcontainers.DockerNetwork,
	runContainer bool,
) (*postgresContainer, error) {
	cfg.ExposedPorts = []string{cfg.Port + "/tcp"}
	if len(cfg.Envs) == 0 {
		cfg.Envs = map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		}
	}

	cr := GetBaseContainerRequest(cfg.BaseContainerConfig, dockerNetwork)

	cr.WaitingFor = wait.ForHTTP(cfg.Port + "/tcp")

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

	return &postgresContainer{
		Container: container,
		URI:       fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres", ip, mappedPort.Port()),
	}, nil
}
