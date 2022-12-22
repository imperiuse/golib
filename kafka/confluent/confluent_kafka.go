package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	jsoniter "github.com/json-iterator/go"

	"go.uber.org/zap"
)

const PartitionAny = kafka.PartitionAny

type (
	Consumer interface {
		Start(context.Context, []TopicSpecification, []string, ConsumeEventFuncGoroutine) error
		Stop() error
	}

	Producer interface {
		Start(context.Context, func(Event)) error
		Stop(int)
		Publish(topic *string, partition int32, key []byte, value []byte, deliveryChan chan Event) error
	}

	Adminer interface {
		CreateTopics(ctx context.Context, topics []TopicSpecification, opt ...CreateTopicsAdminOption) error
		Stop()
	}

	ConsumeEventFuncGoroutine = func(context.Context, *kafka.Consumer)
	ProcessEventFunc          = func(context.Context, *kafka.Consumer, *Message)

	TopicPartition          = kafka.TopicPartition
	TopicSpecification      = kafka.TopicSpecification
	CreateTopicsAdminOption = kafka.CreateTopicsAdminOption
	Event                   = kafka.Event
	Message                 = kafka.Message
)

var GenDefaultConsumeEventFunc = func(
	maxWaitReadTimeout time.Duration,
	processEventFunc ProcessEventFunc,
) ConsumeEventFuncGoroutine {
	return func(ctx context.Context, kafkaConsumer *kafka.Consumer) {
		defer func() {
			err := kafkaConsumer.Close()
			if err != nil {
				fmt.Printf("Err while close consumer: %v\n", err)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			default:
				msg, err := kafkaConsumer.ReadMessage(maxWaitReadTimeout)
				if err != nil {
					// Errors are informational and automatically handled by the consumer
					continue
				}

				processEventFunc(ctx, kafkaConsumer, msg)
			}
		}
	}
}

func GenDefaultProcessMsgWithZapLogger[T any](
	log *zap.Logger,
	businessLogicFunc func(string, any) bool,
) ProcessEventFunc {
	return func(ctx context.Context, kafkaConsumer *kafka.Consumer, msg *Message) {
		key, data, err := UnmarshalKafkaValueToStruct[T](msg)
		if err != nil {
			log.Error("UnmarshalKafkaValueToStruct problem", zap.Error(err))

			return
		}

		if isNeedCommitMsg := businessLogicFunc(key, data); !isNeedCommitMsg {
			return
		}

		_, err = kafkaConsumer.CommitMessage(msg)
		if err != nil {
			log.Error("CommitMessage error", zap.Error(err))

			return
		}
	}
}

func UnmarshalKafkaValueToStruct[T any](msg *Message) (key string, data T, err error) {
	key = string(msg.Key)
	value := msg.Value

	if err = jsoniter.Unmarshal(value, &data); err != nil {
		return key, data, fmt.Errorf("jsoniter.Unmarshal kafka msg from value problem. err: %w", err)
	}

	return key, data, nil
}
