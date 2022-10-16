package environment

import (
	"fmt"

	"github.com/imperiuse/golib/testcontainer"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	NetworkName = "test_id_provisioning_network"

	AppName = "app"

	KafkaImage         = "confluentinc/cp-kafka:7.2.0"
	KafkaBrokerPort    = "9092"
	KafkaClientPort    = "9101"
	KafkaContainerName = "broker"

	ZookeeperImage         = "confluentinc/cp-zookeeper:7.2.0"
	ZooKeeperPort          = "2181"
	ZooKeeperContainerName = "zookeeper"
	ZooTickTime            = "2000"

	PostgresImage         = "postgres:14"
	PostgresPort          = "5432"
	PostgresContainerName = "db"
	PostgresUsername      = "admin"
	PostgresPassword      = "password"
	PostgresDB            = "id-provisioning"
)

var (
	zooCfg = testcontainer.ZookeeperConfig{
		BaseContainerConfig: testcontainer.BaseContainerConfig{
			Image:        ZookeeperImage,
			Port:         ZooKeeperPort,
			Name:         ZooKeeperContainerName,
			ExposedPorts: []string{fmt.Sprintf("0.0.0.0:%[1]s:%[1]s", ZooKeeperPort)},
			Envs: map[string]string{
				"ZOOKEEPER_SERVER_ID":   "1",
				"ZOOKEEPER_SERVERS":     "zoo1:2888:3888",
				"ZOOKEEPER_CLIENT_PORT": ZooKeeperPort,
				"ZOOKEEPER_TICK_TIME":   ZooTickTime,
			},
			WaitingForStrategy: wait.ForLog("binding to port 0.0.0.0/0.0.0.0:2181"),
		},
	}

	kafkaCfg = testcontainer.KafkaConfig{
		BaseContainerConfig: testcontainer.BaseContainerConfig{
			Image: KafkaImage,
			Name:  KafkaContainerName,
			Port:  KafkaClientPort,
			ExposedPorts: []string{
				fmt.Sprintf("0.0.0.0:%[1]s:%[1]s", KafkaClientPort),
				fmt.Sprintf("0.0.0.0:%[1]s:%[1]s", KafkaBrokerPort),
			},
			Envs: map[string]string{
				"KAFKA_BROKER_ID":                      "1",
				"KAFKA_ZOOKEEPER_CONNECT":              ZooKeeperContainerName + ":" + ZooKeeperPort,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
				"KAFKA_ADVERTISED_HOST_NAME":           KafkaContainerName,
				"KAFKA_ADVERTISED_LISTENERS": "PLAINTEXT://" + KafkaContainerName + ":29092,PLAINTEXT_HOST://localhost:" +
					KafkaBrokerPort,
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":            "1",
				"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":            "0",
				"KAFKA_CONFLUENT_LICENSE_TOPIC_REPLICATION_FACTOR":  "1",
				"KAFKA_CONFLUENT_BALANCER_TOPIC_REPLICATION_FACTOR": "1",
				"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":               "1",
				"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR":    "1",
				"KAFKA_AUTO_CREATE_TOPICS.ENABLE":                   "true",
			},
			WaitingForStrategy: wait.ForLog("[KafkaServer id=1] started"),
		},
		ClientPort: KafkaClientPort,
		BrokerPort: KafkaBrokerPort,
	}

	postgresCfg = testcontainer.PostgresConfig{BaseContainerConfig: testcontainer.BaseContainerConfig{
		Name:         PostgresContainerName,
		Image:        PostgresImage,
		Port:         PostgresPort,
		ExposedPorts: []string{fmt.Sprintf("0.0.0.0:%[1]s:%[1]s", PostgresPort)},
		Envs: map[string]string{
			"POSTGRES_USER":     PostgresUsername,
			"POSTGRES_PASSWORD": PostgresPassword,
			"POSTGRES_DB":       PostgresDB,
		},
		WaitingForStrategy: wait.ForLog("database system is ready to accept connections"),
	},
	}
)
