package kafka

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type Publisher struct {
	config   PublisherConfig
	producer sarama.SyncProducer
	admin    sarama.ClusterAdmin
	closed   bool
}

// NewPublisher creates a new Kafka Publisher.
func NewPublisher(config PublisherConfig) (*Publisher, error) {
	config.setDefaults()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(config.Brokers, config.OverwriteSaramaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Kafka producer")
	}

	admin, err := sarama.NewClusterAdmin(config.Brokers, config.OverwriteSaramaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Kafka admin")
	}

	return &Publisher{
		config:   config,
		producer: producer,
		admin:    admin,
	}, nil
}

func (c *PublisherConfig) setDefaults() {
	if c.OverwriteSaramaConfig == nil {
		c.OverwriteSaramaConfig = DefaultSaramaSyncPublisherConfig()
	}
}

func (c PublisherConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("missing brokers")
	}
	if c.Marshaler == nil {
		return errors.New("missing marshaler")
	}

	return nil
}

func DefaultSaramaSyncPublisherConfig() *sarama.Config {
	config := sarama.NewConfig()

	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Version = sarama.V1_0_0_0
	config.Metadata.Retry.Backoff = time.Second * 2
	config.ClientID = "go.frame.publisher"

	return config
}

// Publish publishes message to Kafka.
//
// Publish is blocking and wait for ack from Kafka.
// When one of messages delivery fails - function is interrupted.
func (p *Publisher) Publish(topic string, msgs ...*Message) error {
	if p.closed {
		return errors.New("publisher closed")
	}

	//logFields := make(watermill.LogFields, 4)
	//logFields["topic"] = topic

	for _, msg := range msgs {
		//logFields["message_uuid"] = msg.UUID
		//p.logger.Trace("Sending message to Kafka", logFields)

		kafkaMsg, err := p.config.Marshaler.Marshal(topic, msg)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal message %s", msg.UUID)
		}

		_, _, err = p.producer.SendMessage(kafkaMsg)
		if err != nil {
			return errors.Wrapf(err, "cannot produce message %s", msg.UUID)
		}
		//log.Printf("kafka_partition: %d, kafka_partition_offset: %d", partition, offset)
		//logFields["kafka_partition"] = partition
		//logFields["kafka_partition_offset"] = offset
		//
		//p.logger.Trace("Message sent to Kafka", logFields)
	}

	return nil
}

// ListTopicGroupIds 列出指定topic下所有的group id
func (p *Publisher) ListTopicGroupIds(topic string) ([]string, error) {
	consumerGroups, err := p.admin.ListConsumerGroups()
	if err != nil {
		return nil, err
	}
	groupIds := make([]string, 0)
	for groupId := range consumerGroups {
		groupIds = append(groupIds, groupId)
	}

	return groupIds, nil
}

func (p *Publisher) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true

	if err := p.producer.Close(); err != nil {
		return errors.Wrap(err, "cannot close Kafka producer")
	}

	return nil
}
