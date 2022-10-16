package testcontainer

import (
	"context"
	"fmt"
)

type (
	// KafkaCluster - kafka cluster struct (together kafka container and zookeeper).
	KafkaCluster struct {
		KafkaContainer     Container
		ZookeeperContainer Container

		KafkaURI string

		kafkaCfg KafkaConfig
		zooCfg   ZookeeperConfig
	}

	// KafkaConfig - kafka container config with zookeeper.
	KafkaConfig struct {
		ClientPort string
		BrokerPort string
		BaseContainerConfig
	}
)

// NewKafkaCluster - create new kafka cluster (together kafka container + zookeeper), but do not start it yet.
func NewKafkaCluster(
	ctx context.Context,
	kafkaCfg KafkaConfig,
	zooCfg ZookeeperConfig,
	dockerNetwork *DockerNetwork,
) (*KafkaCluster, error) {
	zookeeperContainer, err := NewZookeeperContainer(ctx, zooCfg, dockerNetwork)
	if err != nil {
		return nil, err
	}

	kafkaContainer, err := NewKafkaContainer(ctx, kafkaCfg, dockerNetwork)
	if err != nil {
		return nil, err
	}

	return &KafkaCluster{
		ZookeeperContainer: zookeeperContainer,
		KafkaContainer:     kafkaContainer,
		KafkaURI:           fmt.Sprintf("%s:%s", kafkaCfg.Name, kafkaCfg.Port),
		kafkaCfg:           kafkaCfg,
		zooCfg:             zooCfg,
	}, nil
}

// NewKafkaContainer - create new kafka container, but do not start it yet.
func NewKafkaContainer(
	ctx context.Context,
	cfg KafkaConfig,
	dockerNetwork *DockerNetwork,
) (Container, error) {
	// creates the kafka container, but do not start it yet
	return NewGenericContainer(ctx, cfg.BaseContainerConfig, dockerNetwork)
}

// Start - start ZookeeperContainer and KafkaContainer.
func (c *KafkaCluster) Start(ctx context.Context) error {
	if err := c.ZookeeperContainer.Start(ctx); err != nil {
		return fmt.Errorf("could not start Zookeeper container. err: %w", err)
	}

	return c.KafkaContainer.Start(ctx)
}

// Terminate - terminate ZookeeperContainer and KafkaContainer.
func (c *KafkaCluster) Terminate(ctx context.Context) error {
	err := c.KafkaContainer.Terminate(ctx)
	err2 := c.ZookeeperContainer.Terminate(ctx)

	if err == nil {
		return err2
	}

	if err2 == nil {
		return err
	}

	return fmt.Errorf("err1: %v, err2: %w", err.Error(), err2)
}
