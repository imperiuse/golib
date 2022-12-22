package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	confluent_kafka "github.com/imperiuse/golib/kafka/confluent"
)

type (
	consumer struct {
		*kafka.Consumer
		configMap   *kafka.ConfigMap
		adminClient confluent_kafka.Adminer
	}
)

func New(configMap *kafka.ConfigMap) confluent_kafka.Consumer {
	return &consumer{
		configMap: configMap,
	}
}

func (c *consumer) Stop() error {
	c.adminClient.Stop()
	return c.Consumer.Close()
}

func (c *consumer) Start(
	ctx context.Context,
	createTopics []kafka.TopicSpecification,
	subTopics []string,
	consFunc confluent_kafka.ConsumeEventFuncGoroutine,
) error {
	if len(createTopics) > 0 {
		var err error
		c.adminClient, err = confluent_kafka.NewAdminKafkaClient(c.configMap)
		if err != nil {
			return fmt.Errorf("could not create kafka admin client. err: %w", err)
		}

		if err = c.adminClient.CreateTopics(ctx, createTopics); err != nil {
			return fmt.Errorf("could not CreateTopics via kafka admin client. err: %w", err)
		}
	}

	kafkaConsumer, err := kafka.NewConsumer(c.configMap)
	if err != nil {
		return fmt.Errorf("could not create new kafka consumer. err: %w", err)
	}

	if err = kafkaConsumer.SubscribeTopics(subTopics, nil); err != nil {
		return fmt.Errorf("could not subscribe to topics: %s. err: %w", subTopics, err)
	}

	go consFunc(ctx, kafkaConsumer)

	return nil
}

func (c *consumer) ReadMessage(timeout time.Duration) (*confluent_kafka.Message, error) {
	return c.Consumer.ReadMessage(timeout)
}
func (c *consumer) CommitMessage(msg *confluent_kafka.Message) ([]confluent_kafka.TopicPartition, error) {
	return c.Consumer.CommitMessage(msg)
}
