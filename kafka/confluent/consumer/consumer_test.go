package consumer

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	lib_confluent_kafka "github.com/imperiuse/golib/kafka/confluent"
	"github.com/imperiuse/golib/kafka/confluent/tests"
)

type ConsumerTestSuite struct {
	tests.Suite
	tests.ContainersEnvironment

	consumer            lib_confluent_kafka.Consumer
	consumerKafkaConfig *kafka.ConfigMap
}

func TestConsumerSuite(t *testing.T) {
	if testing.Short() { // do not run without test environment (this is integration test)
		t.Skip()
	}

	suite.Run(t, new(ConsumerTestSuite))
}

func (s *ConsumerTestSuite) SetupSuite() {
	s.T().Log("> From SetupSuite")

	s.consumerKafkaConfig = &kafka.ConfigMap{
		"bootstrap.servers":     tests.BootstrapKafkaServer,
		"group.id":              "consumer_test_" + uuid.New().String(),
		"auto.offset.reset":     "smallest",
		"retry.backoff.ms":      "250",
		"request.required.acks": 1,
		"enable.auto.commit":    false,
	}
}

func (s *ConsumerTestSuite) SetupTest() {
	s.T().Log(">> From SetupTest")

	s.Suite.Setup() // setup base common suite.

	s.StartPureDockerEnvironment(s.T(), s.Ctx)

	s.consumer = New(s.consumerKafkaConfig)
	require.NotEmpty(s.T(), s.consumer)
}

func (s *ConsumerTestSuite) BeforeTest(_, _ string) {
	s.T().Log(">>> From BeforeTest")
}

func (s *ConsumerTestSuite) AfterTest(_, _ string) {
	s.T().Log(">>> From AfterTest")
}

func (s *ConsumerTestSuite) TearDownTest() {
	s.T().Log(">> From TearDownTest")

	s.consumer.Stop()

	s.FinishedPureDockerEnvironment(s.T(), s.Ctx)

	s.Cancel()
}

func (s *ConsumerTestSuite) TearDownSuite() {
	s.T().Log("> From TearDownSuite")
}

func (s *ConsumerTestSuite) TestNegativeConsumeNotExistKafka() {
	s.T().Log("TestNegativeConsumeNotExistKafka")

	var timeStopped = time.Second
	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.KafkaContainer.Stop(context.Background(), &timeStopped))

	<-time.After(timeStopped)

	const (
		maxMsgReadWaitTimeout = time.Millisecond * 100
	)

	cons := New(s.consumerKafkaConfig)
	require.NotEmpty(s.T(), cons)

	err, errChan := cons.Start(
		s.Ctx,
		nil,
		[]string{"some_not_exist_topic"},
		maxMsgReadWaitTimeout,
		func(k *kafka.Consumer, msg *lib_confluent_kafka.Message) {
			s.T().Fail()
		},
	)

	require.Empty(s.T(), err)
	require.NotNil(s.T(), errChan)

	select {
	case <-time.After(time.Second * 3):
		s.T().Fail()
	case err = <-errChan:
		require.NotEmpty(s.T(), err)
		require.True(s.T(), strings.Contains(err.Error(), "Connection refused"))
	}
}

func (s *ConsumerTestSuite) TestConsumeNotExistsTopic() {
	s.T().Log("TestConsumeNotExistsTopic")

	const (
		maxMsgReadWaitTimeout = time.Millisecond * 100
	)

	err, errChan := s.consumer.Start(
		s.Ctx,
		nil,
		[]string{"some_not_exist_topic"},
		maxMsgReadWaitTimeout,
		func(k *kafka.Consumer, msg *lib_confluent_kafka.Message) {
			s.T().Fail()
		},
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), errChan)

	select {
	case <-time.After(time.Second * 3):
		s.T().Fail()
	case err = <-errChan:
		require.NotEmpty(s.T(), err)
	}
}

func (s *ConsumerTestSuite) TestConsumerNotFailedWhenKafkaRestart() {
	s.T().Log("TestConsumerNotFailedWhenKafkaRestart")

	topicName := "test.topic.consumer.not.failed.kafka"
	const (
		maxMsgReadWaitTimeout = time.Millisecond * 100
		maxFlushMs            = 100
	)

	sendMsgs := []string{"msg_1", "msg_2", "msg_3"}

	var cnt int32

	err, errChan := s.consumer.Start(
		s.Ctx,
		[]kafka.TopicSpecification{
			{
				Topic:             topicName,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		},
		[]string{topicName},
		maxMsgReadWaitTimeout,
		func(k *kafka.Consumer, msg *lib_confluent_kafka.Message) {
			s.T().Logf("Received msg: %+v", msg)
			// Committing messages in all cases to remove them from the queue.
			defer func() {
				_, err := k.CommitMessage(msg)
				require.NoError(s.T(), err)
			}()

			atomic.AddInt32(&cnt, 1) // inc cnt received.
		},
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), errChan)

	<-time.After(time.Second) // w8 some time for create topics.

	simpleProducer := createSimpleKafkaProducer(s.T())

	for _, msg := range sendMsgs {
		err := simpleProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
			Key:            []byte(uuid.New().String()),
		}, nil)
		require.NoError(s.T(), err)
	}

	simpleProducer.Flush(maxFlushMs) // flush == empty buffer -> send all msg asap.

	<-time.After(time.Second * 3)

	simpleProducer.Close()

	s.T().Log("Stop kafka container")

	stopTimeout := time.Second
	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.KafkaContainer.Stop(context.Background(), &stopTimeout))

	select {
	case <-time.After(time.Second * 3):
		s.T().Fail()
	case err = <-errChan:
		require.NotEmpty(s.T(), err)
	}

	s.T().Log("Start kafka container")

	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.KafkaContainer.Start(context.Background()))

	<-time.After(time.Second * 5) // w8 some time for create topics.

	simpleProducer = createSimpleKafkaProducer(s.T())
	defer simpleProducer.Close()

	for _, msg := range sendMsgs {
		err := simpleProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
			Key:            []byte(uuid.New().String()),
		}, nil)
		require.NoError(s.T(), err)
	}

	simpleProducer.Flush(maxFlushMs) // flush == empty buffer -> send all msg asap.

	<-time.After(time.Second * 10)

	require.Equal(s.T(), int32(2*len(sendMsgs)), cnt)
}

func (s *ConsumerTestSuite) TestConsumerNotFailedWhenZookeeperRestart() {
	s.T().Log("TestConsumerNotFailedWhenZookeeperRestart")

	topicName := "test.topic.consumer.not.failed.zookeeper"
	const (
		maxMsgReadWaitTimeout = time.Millisecond * 100
		maxFlushMs            = 100
	)

	sendMsgs := []string{"msg_1", "msg_2", "msg_3"}

	var cnt int32

	err, errChan := s.consumer.Start(
		s.Ctx,
		[]kafka.TopicSpecification{
			{
				Topic:             topicName,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		},
		[]string{topicName},
		maxMsgReadWaitTimeout,
		func(k *kafka.Consumer, msg *lib_confluent_kafka.Message) {
			s.T().Logf("Received msg: %+v", msg)
			// Committing messages in all cases to remove them from the queue.
			defer func() {
				_, err := k.CommitMessage(msg)
				require.NoError(s.T(), err)
			}()

			atomic.AddInt32(&cnt, 1) // inc cnt received.
		},
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), errChan)

	<-time.After(time.Second) // w8 some time for create topics.

	simpleProducer := createSimpleKafkaProducer(s.T())

	for _, msg := range sendMsgs {
		err := simpleProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
			Key:            []byte(uuid.New().String()),
		}, nil)
		require.NoError(s.T(), err)
	}

	simpleProducer.Flush(maxFlushMs) // flush == empty buffer -> send all msg asap.

	<-time.After(time.Second * 3)

	simpleProducer.Close()

	s.T().Log("Stop zookeeper container")

	stopTimeout := time.Second
	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.ZookeeperContainer.Stop(context.Background(), &stopTimeout))

	<-time.After(time.Second * 5)

	s.T().Log("Start zookeeper container")

	require.NoError(s.T(), s.ContainersEnvironment.KafkaCluster.ZookeeperContainer.Start(context.Background()))

	<-time.After(time.Second * 5) // w8 some time for create topics.

	simpleProducer = createSimpleKafkaProducer(s.T())
	defer simpleProducer.Close()

	for _, msg := range sendMsgs {
		err := simpleProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
			Key:            []byte(uuid.New().String()),
		}, nil)
		require.NoError(s.T(), err)
	}

	simpleProducer.Flush(maxFlushMs) // flush == empty buffer -> send all msg asap.

	<-time.After(time.Second * 10)

	require.Equal(s.T(), int32(2*len(sendMsgs)), cnt)
}

func (s *ConsumerTestSuite) TestSuccessConsumeSomeMsgOnePartition() {
	s.T().Log("TestSuccessConsumeSomeMsgOnePartition")

	topicName := "test.success.consume.some.msg.one.partition"
	const (
		maxMsgReadWaitTimeout = time.Millisecond * 100
		maxFlushMs            = 100
	)

	createTopicSpecs := []kafka.TopicSpecification{
		{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	var (
		cnt      int32
		n        int
		sendMsgs = []string{"msg_1", "msg_2", "msg_3"}
	)

	err, errChan := s.consumer.Start(
		s.Ctx,
		createTopicSpecs,
		[]string{topicName},
		maxMsgReadWaitTimeout,
		func(k *kafka.Consumer, msg *lib_confluent_kafka.Message) {
			s.T().Logf("Received msg: %+v", msg)
			// Committing messages in all cases to remove them from the queue.
			defer func() {
				_, err := k.CommitMessage(msg)
				require.NoError(s.T(), err)
			}()

			if strings.EqualFold(string(msg.Value), sendMsgs[n]) { // check order and value together.
				atomic.AddInt32(&cnt, 1) // inc cnt received.
				n++                      // inc awaitable msg index.
			}
		},
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), errChan)

	<-time.After(time.Second) // w8 some time for create topics.

	simpleProducer := createSimpleKafkaProducer(s.T())
	defer simpleProducer.Close()

	for _, msg := range sendMsgs {
		err := simpleProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
			Key:            []byte(uuid.New().String()),
		}, nil)
		require.NoError(s.T(), err)
	}

	simpleProducer.Flush(maxFlushMs) // flush == empty buffer -> send all msg asap.

	<-time.After(time.Second * 3)

	require.Equal(s.T(), int32(3), cnt)
}

func (s *ConsumerTestSuite) TestSuccessConsumeSomeMsgSomePartitions() {
	s.T().Log("TestSuccessConsumeSomeMsgSomePartitions")

	topicName := "test.success.consume.some.msg.some.partitions"
	const (
		cntPartitions         = 3
		maxMsgReadWaitTimeout = time.Millisecond * 100
		maxFlushMs            = 100
	)

	createTopicSpecs := []kafka.TopicSpecification{
		{
			Topic:             topicName,
			NumPartitions:     cntPartitions,
			ReplicationFactor: 1,
		},
	}

	var (
		cnt       int32
		sendMsgs  = make(map[int]string, cntPartitions)
		consumers = make([]lib_confluent_kafka.Consumer, 0, cntPartitions)
	)

	for i := 0; i < cntPartitions; i++ {
		i := i // fix closure problem.

		cons := New(s.consumerKafkaConfig)
		require.NotEmpty(s.T(), cons)

		consumers = append(consumers, cons)

		sendMsgs[i] = fmt.Sprintf("msg_%d", i)

		var oldPartNum int32 = -1

		err, errChan := consumers[i].Start(
			s.Ctx,
			createTopicSpecs,
			[]string{topicName},
			maxMsgReadWaitTimeout,
			func(k *kafka.Consumer, msg *lib_confluent_kafka.Message) {
				s.T().Logf("Part listener: %d. Received msg: %v. %v", i, string(msg.Value), msg.TopicPartition.Partition)
				// Committing messages in all cases to remove them from the queue.
				defer func() {
					_, err := k.CommitMessage(msg)
					require.NoError(s.T(), err)
				}()

				if oldPartNum == -1 {
					oldPartNum = msg.TopicPartition.Partition
				}

				if msg.TopicPartition.Partition == oldPartNum { // always receive the same partition.
					atomic.AddInt32(&cnt, 1) // inc cnt received.
				}
				// consumers start consume random partition! probably it can be configurable, but I have no idea right now. ->
				// DOCS: https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
			},
		)

		require.NoError(s.T(), err)
		require.NotNil(s.T(), errChan)
	}

	<-time.After(time.Second) // w8 some time for create topics.

	simpleProducer := createSimpleKafkaProducer(s.T())
	defer simpleProducer.Close()

	const cntRepeat = 2
	for j := 0; j < cntRepeat; j++ {
		for partitionKey, msg := range sendMsgs {
			err := simpleProducer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: int32(partitionKey)},
				Value:          []byte(msg),
				Key:            []byte(uuid.New().String()),
			}, nil)
			require.NoError(s.T(), err)
		}
	}

	simpleProducer.Flush(maxFlushMs) // flush == empty buffer -> send all msg asap.

	<-time.After(time.Second * 10)

	require.Equal(s.T(), cntRepeat*int32(cntPartitions), cnt)
}

func createSimpleKafkaProducer(t *testing.T) *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": tests.BootstrapKafkaServer})
	require.NoError(t, err)
	require.NotEmpty(t, p)

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) { //nolint:singleCaseSwitch // this is for tests only.
			case *kafka.Message:
				require.NoErrorf(t, ev.TopicPartition.Error, "Delivery failed: %v\n", ev.TopicPartition)
				t.Logf("Delivered message to %v %v %v %v \n",
					ev.TopicPartition, ev.Headers, string(ev.Key), string(ev.Value))
			}
		}
	}()

	return p
}
