package confluent

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
		Start(ctx context.Context, createTopics []TopicSpecification, consumeTopics []string, maxWaitReadTimeout time.Duration, f ProcessEventFunc) (error, ErrChan)
		Stop()
	}

	Producer interface {
		Start(ctx context.Context, finishTimeoutFlushInMs int) (chan Event, error)
		Stop()
		Publish(msg *Message) error
		Flush(timeoutFlushInMs int)
	}

	Adminer interface {
		CreateTopics(ctx context.Context, topics []TopicSpecification, opt ...CreateTopicsAdminOption) error
		Stop()
	}

	ProcessEventFunc = func(*kafka.Consumer, *Message)
	ErrChan          = chan error // read only channel. non block. store error of kafka if not full.

	TopicPartition          = kafka.TopicPartition
	TopicSpecification      = kafka.TopicSpecification
	CreateTopicsAdminOption = kafka.CreateTopicsAdminOption
	Event                   = kafka.Event
	Message                 = kafka.Message
)

func GenDefaultProcessMsgWithZapLogger[T any](
	log *zap.Logger,
	businessLogicFunc func(string, any) bool,
) ProcessEventFunc {
	return func(kafkaConsumer *kafka.Consumer, msg *Message) {
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
