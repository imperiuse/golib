package producer

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"

	confluent_kafka "github.com/imperiuse/golib/kafka/confluent"
)

type (
	confluentKafkaProducer struct {
		*kafka.Producer
		configMap *kafka.ConfigMap
	}
)

var GenDefaultDeliveryHandlerFuncWithZapLogger = func(log *zap.Logger) func(event kafka.Event) {
	return func(e kafka.Event) {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Error(fmt.Sprintf("confluentKafkaProducer: Delivery failed: %v\n", ev.TopicPartition),
					zap.Any("tp", ev.TopicPartition))
				return
			}

			log.Debug(fmt.Sprintf("Successfully produced record to topic %s partition [%d] @ offset %v",
				*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset),
				zap.Any("tp", ev.TopicPartition))

		case *kafka.Error:
			log.Error("confluentKafkaProducer: Received kafka.Error msg", zap.Error(ev))
		default:
			log.Error("confluentKafkaProducer: Received unknown error", zap.Any("ev", ev))
		}
	}
}

// New - "constructor" for kafka confluentKafkaProducer.
func New(configMap *kafka.ConfigMap) confluent_kafka.Producer {
	return &confluentKafkaProducer{
		configMap: configMap,
	}
}

// Start - start async kafka publishing.
func (p *confluentKafkaProducer) Start(ctx context.Context, processProducerEvents func(kafka.Event)) error {
	var err error

	p.Producer, err = kafka.NewProducer(p.configMap)
	if err != nil {
		return fmt.Errorf("confluentKafkaProducer.Start: could not create new kafka confluentKafkaProducer. err: %w", err)
	}

	// Delivery report handler for produced messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case e := <-p.Producer.Events():
				processProducerEvents(e)
			}
		}
	}()

	return nil
}

// Stop - stop async publishing.
func (p *confluentKafkaProducer) Stop(timeoutFlushInMs int) {
	// Wait for message deliveries before shutting down
	p.Flush(timeoutFlushInMs)

	p.Producer.Close()
}

// Publish - publish.
func (p *confluentKafkaProducer) Publish(
	topic *string,
	partition int32,
	key []byte,
	value []byte,
	deliveryChan chan confluent_kafka.Event,
) error {
	// Produce messages to topic (asynchronously)
	err := p.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: topic, Partition: partition},
			Value:          value,
			Key:            key,
		}, deliveryChan)
	if err != nil {
		return fmt.Errorf("confluentKafkaProducer.Publish: could not Produce data. "+
			"err: %w. topic: %v, partition: %v, key: %v", err, topic, partition, key)
	}

	return nil
}
