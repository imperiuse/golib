package environment

import (
	"context"
	"testing"
	"time"

	"github.com/imperiuse/golib/testcontainer"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ContainersEnvironment struct {
	// First - native go-testcontainers way for creating docker containers.
	dockerNetwork     *testcontainers.DockerNetwork
	kafkaContainer    *testcontainer.KafkaCluster
	postgresContainer testcontainers.Container

	// Second - docker-compose way + go-testcontainers for create docker container env.
	compose testcontainers.DockerCompose
}

// StartPureDockerEnvironment - create and start docker containers env with first way.
func (c *ContainersEnvironment) StartPureDockerEnvironment(t *testing.T, ctx context.Context) {
	t.Log("> From SetupEnvironment")

	t.Log("Create docker network")
	dn, err := testcontainer.NewDockerNetwork(ctx, NetworkName)
	require.Nil(t, err, "error must be nil for NewDockerNetwork")
	require.NotNil(t, dn, "docker network must be not nil")
	c.dockerNetwork = dn.(*testcontainers.DockerNetwork)

	t.Log("Create service deps")
	c.kafkaContainer, err = testcontainer.NewKafkaCluster(ctx, kafkaCfg, zooCfg, c.dockerNetwork)
	require.Nil(t, err, "error must be nil, when create NewKafkaCluster")
	require.NotNil(t, c.kafkaContainer, "kafka cluster must be not nil")

	c.postgresContainer, err = testcontainer.NewPostgresContainer(ctx, postgresCfg, c.dockerNetwork)
	require.Nil(t, err, "error must be nil, when create NewPostgresContainer")
	require.NotNil(t, c.postgresContainer, "postgres container must be not nil")

	t.Log("Start deps services containers")

	require.Nil(t, c.kafkaContainer.Start(ctx), "kafka cluster must start without errors")

	require.Nil(t, c.postgresContainer.Start(ctx), "postgres must start without errors")

	const magicTime = time.Second * 3
	time.Sleep(magicTime) // time sleep development // todo think how to remove this
}

// FinishedPureDockerEnvironment - finished containers (env) which we created by first way.
func (c *ContainersEnvironment) FinishedPureDockerEnvironment(t *testing.T, ctx context.Context) {
	require.Nil(t, testcontainer.TerminateIfNotNil(ctx, c.kafkaContainer), "must not get an error while terminate kafka cluster")
	require.Nil(t, testcontainer.TerminateIfNotNil(ctx, c.postgresContainer), "must not get an error while terminate postgres cluster")
	require.Nil(t, c.dockerNetwork.Remove(ctx), "must not get an error while remove docker network")
}

// StartDockerComposeEnvironment - create and start docker containers env with second way.
func (c *ContainersEnvironment) StartDockerComposeEnvironment(
	t *testing.T,
	composeFilePaths []string,
	identifier string,
) {
	c.compose = testcontainers.NewLocalDockerCompose(composeFilePaths, identifier).
		WaitForService(PostgresContainerName, wait.ForLog("database system is ready to accept connections")).
		WaitForService(ZooKeeperContainerName, wait.ForLog("binding to port 0.0.0.0/0.0.0.0:"+ZooKeeperPort)).
		WaitForService(KafkaContainerName, wait.ForLog("[KafkaServer id=1] started"))

	if len(composeFilePaths) > 1 { // this is little tricky hack here. :)
		// if we have one docker-compose file for app container, that add wait strategy.
		c.compose = c.compose.WaitForService(AppName, wait.ForLog("App starting successfully! Ready for hard work!"))
	}

	require.Nil(t, c.compose.WithCommand([]string{"up", "--force-recreate", "-d"}).Invoke().Error)
}

// FinishedDockerComposeEnvironment - finished containers (env) which we created by second way.
func (c *ContainersEnvironment) FinishedDockerComposeEnvironment(t *testing.T) {
	require.Nil(t, c.compose.Down().Error, "docker compose must down without errors")
}
