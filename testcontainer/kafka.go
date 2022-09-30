package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

type (
	// KafkaCluster - kafka cluster struct (together kafka container and zookeeper).
	KafkaCluster struct {
		KafkaContainer     testcontainers.Container
		ZookeeperContainer testcontainers.Container

		kafkaCfg KafkaConfig
		zooCfg   ZookeeperConfig
	}

	// KafkaConfig - kafka container config with zookeeper.
	KafkaConfig struct {
		BaseContainerConfig
		ClientPort string
		BrokerPort string
	}
)

// NewKafkaCluster - create new kafka cluster (together kafka container + zookeeper), but do not start it yet.
func NewKafkaCluster(
	ctx context.Context,
	kafkaCfg KafkaConfig,
	zooCfg ZookeeperConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) (*KafkaCluster, error) {
	zookeeperContainer, err := NewZookeeperContainer(ctx, zooCfg, dockerNetwork)
	if err != nil {
		return nil, err
	}

	kafkaContainer, err := NewKafkaContainer(ctx, kafkaCfg, zooCfg, dockerNetwork)
	if err != nil {
		return nil, err
	}

	return &KafkaCluster{
		ZookeeperContainer: zookeeperContainer,
		KafkaContainer:     kafkaContainer,
		kafkaCfg:           kafkaCfg,
		zooCfg:             zooCfg,
	}, nil
}

// NewKafkaContainer - create new kafka container, but do not start it yet.
func NewKafkaContainer(
	ctx context.Context,
	kafkaCfg KafkaConfig,
	zooCfg ZookeeperConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) (testcontainers.Container, error) {
	if len(kafkaCfg.ExposedPorts) == 0 {
		kafkaCfg.ExposedPorts = []string{kafkaCfg.ClientPort}
	}

	if len(kafkaCfg.Envs) == 0 {
		kafkaCfg.Envs = map[string]string{
			"KAFKA_BROKER_ID":                      "1",
			"KAFKA_ZOOKEEPER_CONNECT":              "zookeeper:" + zooCfg.Port,
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS": "PLAINTEXT://" + kafkaCfg.Name + ":29092,PLAINTEXT_HOST://localhost:" +
				kafkaCfg.BrokerPort,
			"KAFKA_METRIC_REPORTERS":                            "io.confluent.metrics.reporter.ConfluentMetricsReporter",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":            "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":            "0",
			"KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR":  "1",
			"KAFKA_CONFLUENT_BALANCER_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":               "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR":    "1",
			"KAFKA_JMX_PORT":                               "9101",
			"KAFKA_JMX_HOSTNAME":                           "localhost",
			"KAFKA_CONFLUENT_SCHEMA_REGISTRY_URL":          "http://schema-registry:8089",
			"CONFLUENT_METRICS_REPORTER_BOOTSTRAP_SERVERS": kafkaCfg.Name + ":29092",
			"CONFLUENT_METRICS_REPORTER_TOPIC_REPLICAS":    "1",
			"CONFLUENT_METRICS_ENABLE":                     "true",
			"CONFLUENT_SUPPORT_CUSTOMER_ID":                "anonymous",
			"KAFKA_AUTO_CREATE_TOPICS.ENABLE":              "true",
		}
	}

	// creates the kafka container, but do not start it yet
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: GetBaseContainerRequest(kafkaCfg.BaseContainerConfig, dockerNetwork),
	})
}

// Start - start ZookeeperContainer and KafkaContainer.
func (c *KafkaCluster) Start(ctx context.Context) error {
	if err := c.ZookeeperContainer.Start(ctx); err != nil {
		return fmt.Errorf("could not start container. err: %w", err)
	}

	const waitZoo = 5 * time.Second
	time.Sleep(waitZoo) // time sleep development

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
