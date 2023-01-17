package confluent

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type (
	AdminKafkaClient struct {
		*kafka.AdminClient
	}
)

func NewAdminKafkaClient(kafkaConfig *kafka.ConfigMap) (Adminer, error) {
	adminClient, err := kafka.NewAdminClient(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create new AdminKafkaClient. err: %w", err)
	}

	return &AdminKafkaClient{
		AdminClient: adminClient,
	}, nil
}

func (c *AdminKafkaClient) CreateTopics(
	ctx context.Context,
	topics []kafka.TopicSpecification,
	opt ...kafka.CreateTopicsAdminOption,
) error {
	_, err := c.AdminClient.CreateTopics(ctx, topics, opt...)
	if err != nil {
		return fmt.Errorf("AdminKafkaClient.CreateTopics err: %w", err)
	}

	return nil
}

func (c *AdminKafkaClient) Stop() {
	c.AdminClient.Close()
}
