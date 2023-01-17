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
		configMap    *kafka.ConfigMap
		ctx          context.Context
		cancel       context.CancelFunc
		deliveryChan chan kafka.Event
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
	ctx, cancel := context.WithCancel(context.Background())
	return &confluentKafkaProducer{
		ctx:       ctx,
		cancel:    cancel,
		configMap: configMap,
	}
}

// Start - start async kafka publishing.
func (p *confluentKafkaProducer) Start(ctx context.Context, finishTimeoutFlushInMs int) (chan kafka.Event, error) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.deliveryChan = make(chan kafka.Event)

	var err error
	p.Producer, err = kafka.NewProducer(p.configMap)
	if err != nil {
		return p.deliveryChan, fmt.Errorf("confluentKafkaProducer.Start: could not create new kafka confluentKafkaProducer. err: %w", err)
	}

	go func() {
		defer func() {
			p.Flush(finishTimeoutFlushInMs)
			p.Producer.Close()
			close(p.deliveryChan)
		}()

		for {
			select {
			case <-p.ctx.Done():
				return
			}
		}
	}()

	return p.deliveryChan, nil
}

// Stop - stop async publishing.
func (p *confluentKafkaProducer) Stop() {
	p.cancel()
}

// Flush - flush all msg from buffer.
func (p *confluentKafkaProducer) Flush(timeoutFlushInMs int) {
	// Flush and wait for outstanding messages and requests to complete delivery.
	p.Producer.Flush(timeoutFlushInMs)
}

// Publish - publish report log to reports-service.
func (p *confluentKafkaProducer) Publish(
	msg *confluent_kafka.Message,
) error {
	if err := p.Produce(msg, p.deliveryChan); err != nil {
		return fmt.Errorf("confluentKafkaProducer.Publish: could not Produce data. "+
			"err: %w. topic: %v, partition: %v, key: %s", err, msg.TopicPartition.Topic, msg.TopicPartition.Partition, string(msg.Key))
	}

	return nil
}
