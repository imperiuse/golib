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
		configMap *kafka.ConfigMap

		ctx    context.Context
		cancel context.CancelFunc
	}
)

func New(configMap *kafka.ConfigMap) confluent_kafka.Consumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &consumer{
		ctx:       ctx,
		cancel:    cancel,
		configMap: configMap,
	}
}

func (c *consumer) Stop() {
	c.cancel()
}

func (c *consumer) Start(
	ctx context.Context,
	createTopics []kafka.TopicSpecification,
	subTopics []string,
	maxWaitReadTimeout time.Duration,
	consumeEventFunc confluent_kafka.ProcessEventFunc,
) (error, confluent_kafka.ErrChan) {
	c.ctx, c.cancel = context.WithCancel(ctx)

	errChan := make(confluent_kafka.ErrChan)

	if len(createTopics) > 0 {
		adminClient, err := confluent_kafka.NewAdminKafkaClient(c.configMap)
		if err != nil {
			return fmt.Errorf("could not create kafka admin client. err: %w", err), errChan
		}

		defer adminClient.Stop()

		if err = adminClient.CreateTopics(c.ctx, createTopics); err != nil {
			return fmt.Errorf("could not CreateTopics via kafka admin client. err: %w", err), errChan
		}
	}

	kafkaConsumer, err := kafka.NewConsumer(c.configMap)
	if err != nil || kafkaConsumer == nil {
		return fmt.Errorf("could not create new kafka consumer. err: %w", err), errChan
	}

	if err = kafkaConsumer.SubscribeTopics(subTopics, nil); err != nil {
		return fmt.Errorf("could not subscribe to topics: %s. err: %w", subTopics, err), errChan
	}

	go func() {
		defer func() {
			if err = kafkaConsumer.Close(); err != nil {
				fmt.Printf("c.kafkaConsumer.Close(): %v\n", err)
			}

			close(errChan)
		}()

		for {
			select {
			case <-ctx.Done():
				return

			default:
				msg, err := kafkaConsumer.ReadMessage(maxWaitReadTimeout)
				if err != nil {
					// non-blocking write to the chan, if chan full - drop error.
					select {
					case errChan <- err:
					default:
					}
					continue
				}

				consumeEventFunc(kafkaConsumer, msg)
			}
		}
	}()

	return nil, errChan
}
