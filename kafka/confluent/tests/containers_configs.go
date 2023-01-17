package tests

import (
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/imperiuse/golib/testcontainer"
)

const (
	NetworkName = "test_go_replicant_kafka_confluent"

	KafkaImage              = "confluentinc/cp-kafka:7.2.0"
	KafkaDockerInternalPort = "29092" // https://www.confluent.io/blog/kafka-listeners-explained/ -> @see HOW TO: Connecting to Kafka on Docker
	KafkaHostExternalPort   = "9092"
	KafkaJMXClientPort      = "9999"
	KafkaContainerName      = "broker"

	ZookeeperImage         = "confluentinc/cp-zookeeper:7.2.0"
	ZooKeeperPort          = "2181"
	ZooKeeperContainerName = "zookeeper"
	ZooTickTime            = "2000"

	BootstrapKafkaServer = "localhost:" + KafkaHostExternalPort
)

var (
	pollInterval = time.Millisecond * 100

	zooCfg = testcontainer.ZookeeperConfig{
		BaseContainerConfig: testcontainer.BaseContainerConfig{
			Image:        ZookeeperImage,
			Port:         ZooKeeperPort,
			Name:         ZooKeeperContainerName,
			ExposedPorts: []string{fmt.Sprintf("0.0.0.0:%[1]s:%[1]s", ZooKeeperPort)},
			Envs: map[string]string{
				"ZOOKEEPER_SERVER_ID":   "1",
				"ZOOKEEPER_CLIENT_PORT": ZooKeeperPort,
				"ZOOKEEPER_TICK_TIME":   ZooTickTime,
			},
			WaitingForStrategy: wait.ForLog("binding to port 0.0.0.0/0.0.0.0:2181").WithPollInterval(pollInterval),
		},
	}

	kafkaCfg = testcontainer.KafkaConfig{
		BaseContainerConfig: testcontainer.BaseContainerConfig{
			Image: KafkaImage,
			Port:  KafkaHostExternalPort,
			ExposedPorts: []string{
				testcontainer.DoubledPort(KafkaHostExternalPort),
				testcontainer.DoubledPort(KafkaDockerInternalPort),
				testcontainer.DoubledPort(KafkaJMXClientPort),
			},
			Envs: map[string]string{
				"KAFKA_BROKER_ID":                      "1",
				"KAFKA_ZOOKEEPER_CONNECT":              "zookeeper:" + zooCfg.Port,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "DOCKER:PLAINTEXT,HOST:PLAINTEXT",
				"KAFKA_ADVERTISED_LISTENERS": fmt.Sprintf(
					"DOCKER://%s:%s,HOST://localhost:%s",
					KafkaContainerName, KafkaDockerInternalPort,
					KafkaHostExternalPort), // https://www.confluent.io/blog/kafka-listeners-explained/ // Connecting to Kafka on Docker
				"KAFKA_INTER_BROKER_LISTENER_NAME":                  "HOST",
				"KAFKA_JMX_PORT":                                    KafkaJMXClientPort,
				"KAFKA_JMX_HOSTNAME":                                "localhost",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":            "1",
				"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":            "0",
				"KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR":  "1",
				"KAFKA_CONFLUENT_BALANCER_TOPIC_REPLICATION_FACTOR": "1",
				"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":               "1",
				"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR":    "1",
				"KAFKA_AUTO_CREATE_TOPICS.ENABLE":                   "true",
			},
			WaitingForStrategy: wait.ForLog("[KafkaServer id=1] started").WithPollInterval(pollInterval).WithPollInterval(pollInterval),
		},
	}
)
