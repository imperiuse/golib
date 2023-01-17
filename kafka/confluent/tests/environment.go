package tests

import (
	"context"
	"testing"

	"github.com/imperiuse/golib/testcontainer"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

type ContainersEnvironment struct {
	dockerNetwork *testcontainers.DockerNetwork
	KafkaCluster  *testcontainer.KafkaCluster
}

// StartPureDockerEnvironment - create and start docker containers env with first way.
func (c *ContainersEnvironment) StartPureDockerEnvironment(t *testing.T, ctx context.Context) {
	t.Log("Start Test Pure Docker based Environment")

	t.Log("Create docker network")
	dn, err := testcontainer.NewDockerNetwork(ctx, NetworkName)
	require.Nil(t, err, "error must be nil for NewDockerNetwork")
	require.NotNil(t, dn, "docker network must be not nil")
	c.dockerNetwork = dn.(*testcontainers.DockerNetwork)

	t.Log("Create service deps")
	c.KafkaCluster, err = testcontainer.NewKafkaCluster(ctx, kafkaCfg, zooCfg, c.dockerNetwork)
	require.Nil(t, err, "error must be nil, when create NewKafkaCluster")
	require.NotNil(t, c.KafkaCluster, "kafka cluster must be not nil")

	require.Nil(t, c.KafkaCluster.Start(ctx), "kafka cluster must start without errors")
}

// FinishedPureDockerEnvironment - finished containers (env) which we created by first way.
func (c *ContainersEnvironment) FinishedPureDockerEnvironment(t *testing.T, ctx context.Context) {
	t.Log("Finished Test Pure Docker Environment from files")
	require.Nil(t, testcontainer.TerminateIfNotNil(ctx, c.KafkaCluster), "must not get an error while terminate kafka cluster")
	require.Nil(t, c.dockerNetwork.Remove(ctx), "must not get an error while remove docker network")
}
