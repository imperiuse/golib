package producer

import (
	"context"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	confluent_kafka "github.com/imperiuse/golib/kafka/confluent"
	"github.com/imperiuse/golib/kafka/confluent/tests"
)

type ProducerTestSuite struct {
	tests.Suite
	tests.ContainersEnvironment

	producer            confluent_kafka.Producer
	producerKafkaConfig *kafka.ConfigMap
}

func TestProducerSuite(t *testing.T) {
	if testing.Short() { // do not run without test environment (this is integration test)
		t.Skip()
	}

	suite.Run(t, new(ProducerTestSuite))
}

func (s *ProducerTestSuite) SetupSuite() {
	s.T().Log("> From SetupSuite")

	s.producerKafkaConfig = &kafka.ConfigMap{
		"bootstrap.servers": tests.BootstrapKafkaServer,
		"group.id":          "producer_test_" + uuid.New().String(),
	}
}

func (s *ProducerTestSuite) SetupTest() {
	s.T().Log(">> From SetupTest")

	s.Suite.Setup() // setup base common suite.

	s.StartPureDockerEnvironment(s.T(), s.Ctx)

	s.producer = New(s.producerKafkaConfig)
	require.NotEmpty(s.T(), s.producer)
}

func (s *ProducerTestSuite) BeforeTest(_, _ string) {
	s.T().Log(">>> From BeforeTest")
}

func (s *ProducerTestSuite) AfterTest(_, _ string) {
	s.T().Log(">>> From AfterTest")
}

func (s *ProducerTestSuite) TearDownTest() {
	s.T().Log(">> From TearDownTest")

	s.producer.Stop()

	s.FinishedPureDockerEnvironment(s.T(), s.Ctx)

	s.Cancel()
}

func (s *ProducerTestSuite) TearDownSuite() {
	s.T().Log("> From TearDownSuite")
}

func (s *ProducerTestSuite) TestNegativeProduceNotExistKafka() {
	s.T().Log("TestNegativeProduceNotExistKafka")

	var timeStopped = time.Second
	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.KafkaContainer.Stop(context.Background(), &timeStopped))

	<-time.After(timeStopped)

	p := New(s.producerKafkaConfig)
	require.NotEmpty(s.T(), p)

	errChan, err := p.Start(
		s.Ctx,
		100,
	)

	require.Nil(s.T(), err, "err == %v", err)
	require.NotNil(s.T(), errChan)

	topic := "topic"
	require.NoError(s.T(), p.Publish(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(`123`),
		Key:   []byte(`123`),
	}))

	select {
	case err := <-errChan:
		println(err.String())
		return
	case <-time.After(time.Second * 3):
		// todo check this place more.
		//s.T().Log("Not received event about bad delivery")
		//s.T().Fail()
	}
}

func (s *ProducerTestSuite) TestProducerNotFailedWhenKafkaRestart() {
	s.T().Log("TestProducerNotFailedWhenKafkaRestart")

	const timeoutMs = 100

	var (
		topic       = "test.topic.produce.not.failed.kafka.restart"
		msgs        = []string{"msg_0", "msg_1", "msg_2"}
		cntSent     int
		cntReceived int
	)

	deliveryChan, err := s.producer.Start(s.Ctx, timeoutMs)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), deliveryChan)

	for _, msg := range msgs {
		require.NoError(s.T(),
			s.producer.Publish(
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
					Value:          []byte(msg),
					Key:            []byte(uuid.New().String()),
				}))
	}
	s.producer.Flush(timeoutMs)

	// Delivery report handler for produced messages
	go func() {
		for e := range deliveryChan {
			switch ev := e.(type) { //nolint:singleCaseSwitch // this is example only
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					s.T().Logf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					cntSent++
					s.T().Logf("Delivered message to %v %v %v %v \n",
						ev.TopicPartition, ev.Headers, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	require.NoError(s.T(), s.startSimpleConsumer(
		topic,
		func(message *confluent_kafka.Message, err error) {
			if err == nil {
				cntReceived++
			}
		},
	),
	)

	<-time.After(time.Second * 3)

	require.Equal(s.T(), len(msgs), cntSent)
	require.Equal(s.T(), len(msgs), cntReceived)

	s.T().Log("Stop kafka container")

	stopTimeout := time.Second
	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.KafkaContainer.Stop(context.Background(), &stopTimeout))

	<-time.After(time.Second * 3)

	s.T().Log("Start kafka container")

	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.KafkaContainer.Start(context.Background()))

	<-time.After(time.Second * 1) // w8 some time for create topics.

	for _, msg := range msgs {
		require.NoError(s.T(),
			s.producer.Publish(
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
					Value:          []byte(msg),
					Key:            []byte(uuid.New().String()),
				}))
	}
	s.producer.Flush(timeoutMs)

	<-time.After(time.Second * 3)

	require.Equal(s.T(), len(msgs), cntSent)
	require.Equal(s.T(), len(msgs), cntReceived)
}

func (s *ProducerTestSuite) TestSuccessProduceSomeMsgOnePartitionTopic() {
	s.T().Log("From TestSuccessProduceSomeMsgOnePartitionTopic")

	const timeoutMs = 100

	var (
		topic       = "test.topic.success.produce.some.msg.one.partition"
		msgs        = []string{"msg_0", "msg_1", "msg_2"}
		cntSent     int
		cntReceived int
	)

	deliveryChan, err := s.producer.Start(s.Ctx, timeoutMs)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), deliveryChan)

	for _, msg := range msgs {
		require.NoError(s.T(),
			s.producer.Publish(
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
					Value:          []byte(msg),
					Key:            []byte(uuid.New().String()),
				}))
	}
	s.producer.Flush(timeoutMs)

	// Delivery report handler for produced messages
	go func() {
		for e := range deliveryChan {
			switch ev := e.(type) { //nolint:singleCaseSwitch // this is example only
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					s.T().Logf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					cntSent++
					s.T().Logf("Delivered message to %v %v %v %v \n",
						ev.TopicPartition, ev.Headers, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	require.NoError(s.T(), s.startSimpleConsumer(
		topic,
		func(message *confluent_kafka.Message, err error) {
			if err == nil {
				cntReceived++
			}
		},
	),
	)

	<-time.After(time.Second * 3)

	require.Equal(s.T(), len(msgs), cntSent)
	require.Equal(s.T(), len(msgs), cntReceived)
}

func (s *ProducerTestSuite) TestFailedProduceSomeMsgOnePartitionTopic() {
	s.T().Log("From TestFailedProduceSomeMsgOnePartitionTopic")

	const timeoutMs = 100

	var (
		topic        = "test.topic.failed.produce.some.msg.one.partition"
		msgs         = map[int]string{0: "msg_0", 1: "msg_1", 2: "msg_2"}
		cntSent      int
		cntDelivered int
		cntReceived  int
		cntFailed    int
	)

	deliveryChan, err := s.producer.Start(s.Ctx, timeoutMs)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), deliveryChan)

	for partitionKey, msg := range msgs {
		require.NoError(s.T(),
			s.producer.Publish(
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: int32(partitionKey)},
					Value:          []byte(msg),
					Key:            []byte(uuid.New().String()),
				}))
		cntSent++
	}
	s.producer.Flush(timeoutMs)

	// Delivery report handler for produced messages
	go func() {
		for e := range deliveryChan {
			switch ev := e.(type) { //nolint:singleCaseSwitch // this is example only
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					s.T().Logf("Delivery failed: %v\n", ev.TopicPartition)
					cntFailed++
				} else {
					cntDelivered++
					s.T().Logf("Delivered message to %v %v %v %v \n",
						ev.TopicPartition, ev.Headers, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	require.NoError(s.T(), s.startSimpleConsumer(
		topic,
		func(message *confluent_kafka.Message, err error) {
			if err == nil {
				cntReceived++
			}
		},
	),
	)

	<-time.After(time.Second * 3)

	require.Equal(s.T(), len(msgs), cntSent)
	require.Equal(s.T(), 1, cntDelivered)
	require.Equal(s.T(), 1, cntReceived)
	require.Equal(s.T(), 2, cntFailed) // must be 2 error, ->
	// Delivery failed: test.topic.success.produce.some.msg.many.partition[1]@unset(Local: Unknown partition)
	// it's good, we have auto create topic with only one partition.
}

func (s *ProducerTestSuite) startSimpleConsumer(topicName string, consumeFunc func(*confluent_kafka.Message, error)) error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  tests.BootstrapKafkaServer,
		"group.id":           "producer_test_" + uuid.New().String(),
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	})
	if err != nil {
		return err
	}

	err = c.SubscribeTopics([]string{topicName}, nil)
	if err != nil {
		return err
	}

	go func() {
		defer func() { _ = c.Close() }()
		for {
			select {
			case <-s.Ctx.Done():
				return
			default:
				msg, err := c.ReadMessage(-1)
				if err == nil {
					s.T().Logf("Message on %s: %s    \t Key: %s \n", msg.TopicPartition, string(msg.Value), string(msg.Key))
				} else {
					// The client will automatically try to recover from all errors.
					s.T().Logf("Consumer error: %v (%v)\n", err, msg)
				}

				consumeFunc(msg, err)
			}
		}
	}()

	return nil
}
