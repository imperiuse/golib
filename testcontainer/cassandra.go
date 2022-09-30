package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

type (
	// CassandraCluster - cassandra cluster struct.
	CassandraCluster struct {
		cassandra  testcontainers.Container
		queryCQLSH testcontainers.Container
		migrations testcontainers.Container

		cfg CassandraConfig
	}

	// CassandraConfig - cassandra container config.
	CassandraConfig struct { //nolint: fieldalignment,gocritic,govet // this is tests
		BaseContainerConfig

		ContainerWarmUpTimeout  time.Duration
		BeforeMigrationsTimeout time.Duration
		AfterMigrationsTimeout  time.Duration

		CqlshImage             string
		PathToCqlshQueryScript string

		MigrationImage      string
		PathToMigrationsDir string
	}
)

// NewCassandraCluster -create new CassandraCluster.
func NewCassandraCluster(
	ctx context.Context,
	cfg CassandraConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) (*CassandraCluster, error) {
	cassandra, err := NewCassandraContainer(ctx, cfg, dockerNetwork)
	if err != nil {
		return nil, fmt.Errorf("could not create cassandra container: %w", err)
	}

	cqlsh, err := NewCQLSHContainer(ctx, cfg, dockerNetwork, cfg.Name, cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("could not create cql sh container: %w", err)
	}

	migrations, err := NewMigrationsContainer(ctx, cfg, dockerNetwork, cfg.Name, cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("could not create migration container: %w", err)
	}

	return &CassandraCluster{cassandra: cassandra, queryCQLSH: cqlsh, migrations: migrations, cfg: cfg}, nil
}

// Start - start cassandra cluster with necessary migrations.
func (c *CassandraCluster) Start(ctx context.Context) error {
	// 1. Start cassandra container
	if err := c.cassandra.Start(ctx); err != nil {
		return fmt.Errorf("could not start cassndra container: %w", err)
	}

	// time sleep driven development, wait at least 30 sec for warm up cassandra
	time.Sleep(c.cfg.ContainerWarmUpTimeout)

	// 2. Create keyspace
	if err := c.queryCQLSH.Start(ctx); err != nil {
		return fmt.Errorf("could not start queryCQLSH container: %w", err)
	}

	// time sleep driven development
	time.Sleep(c.cfg.BeforeMigrationsTimeout)

	// 3. Apply migrations
	if err := c.migrations.Start(ctx); err != nil {
		return fmt.Errorf("could not start migrations container: %w", err)
	}

	// time sleep driven development
	time.Sleep(c.cfg.AfterMigrationsTimeout)

	return nil
}

// Terminate - terminate container with cassandra.
func (c *CassandraCluster) Terminate(ctx context.Context) error {
	_ = c.queryCQLSH.Terminate(ctx)
	_ = c.migrations.Terminate(ctx)

	return c.cassandra.Terminate(ctx)
}

// NewCassandraContainer - create cassandra container, but do not start it yet.
func NewCassandraContainer(
	ctx context.Context,
	cfg CassandraConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) (testcontainers.Container, error) {
	cfg.ExposedPorts = []string{cfg.Port}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: GetBaseContainerRequest(cfg.BaseContainerConfig, dockerNetwork),
	})
}

// NewCQLSHContainer - create docker container with cql sh (for before migration query script set up).
func NewCQLSHContainer(
	ctx context.Context,
	cfg CassandraConfig,
	dockerNetwork *testcontainers.DockerNetwork,
	cassandraHost string,
	cassandraPort string,
) (testcontainers.Container, error) {
	cr := GetBaseContainerRequest(BaseContainerConfig{
		Name:         "Cqlsh",
		Image:        cfg.CqlshImage,
		Port:         "",
		ExposedPorts: []string{},
		Envs: map[string]string{
			"CQLSH_HOST": cassandraHost,
			"CQLSH_PORT": cassandraPort,
			"CQLVERSION": "3.4.5",
		},
	},
		dockerNetwork,
	)

	cr.Mounts = testcontainers.Mounts(
		testcontainers.BindMount(
			cfg.PathToCqlshQueryScript,
			"/scripts/data.cql",
		),
	)
	cr.AutoRemove = true

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: cr,
	})
}

// NewMigrationsContainer - create docker container for migrations.
func NewMigrationsContainer(
	ctx context.Context,
	cfg CassandraConfig,
	dockerNetwork *testcontainers.DockerNetwork,
	cassandraHost string,
	cassandraPort string,
) (testcontainers.Container, error) {
	cr := GetBaseContainerRequest(BaseContainerConfig{
		Name:         "Cqlsh",
		Image:        cfg.MigrationImage,
		Port:         "",
		ExposedPorts: []string{},
		Envs: map[string]string{
			"CQLSH_HOST": cassandraHost,
			"CQLSH_PORT": cassandraPort,
			"CQLVERSION": "3.4.5",
		},
	},
		dockerNetwork,
	)

	cr.Mounts = testcontainers.Mounts(
		testcontainers.BindMount(
			cfg.PathToMigrationsDir,
			"/migrations",
		),
	)

	cr.Cmd = []string{
		fmt.Sprintf("-source file://migrations -database cassandra://%s:%s/user_properties up",
			cassandraHost, cassandraPort)}

	cr.AutoRemove = true

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: cr,
	})
}
