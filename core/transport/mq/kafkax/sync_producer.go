package kafkax

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"log"
	"time"
)

type syncProducer struct {
	producer sarama.SyncProducer
	admin    sarama.ClusterAdmin
}

func NewSyncProducer(c *ProducerConfig) (p Producer, err error) {
	var producer sarama.SyncProducer
	addr := []string{c.Addr}
	sconf := NewSaramaConfig(c)
	producer, err = sarama.NewSyncProducer(addr, sconf)
	if err != nil {
		fmt.Printf("create sync producer error %v", err.Error())
		return nil, err
	}
	admin, err := sarama.NewClusterAdmin([]string{c.Addr}, sconf)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Kafka admin")
	}
	p = &syncProducer{producer: producer, admin: admin}
	return p, err
}

func NewSaramaConfig(c *ProducerConfig) *sarama.Config {
	conf := sarama.NewConfig()
	conf.Producer.Timeout = 100 * time.Millisecond
	if c.Mechanism != "" {
		conf.Net.SASL.Enable = true
		conf.Net.SASL.Mechanism = sarama.SASLMechanism(c.Mechanism)
		conf.Net.SASL.User = c.UserName
		conf.Net.SASL.Password = c.Password
		conf.Net.SASL.Handshake = true
	}
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	return conf
}

// ListTopicGroupIds 列出指定topic下所有的group id
func (p *syncProducer) ListTopicGroupIds(topic string) ([]string, error) {
	groupIds := make([]string, 0)
	//查询该topic下面的所有的新增groupID
	consumerGroups, err := p.admin.ListConsumerGroups()
	if err != nil {
		log.Printf("failed to list consumer groups, err info: %v", err.Error())
		return nil, err
	}
	gs := make([]string, 0)
	for g, _ := range consumerGroups {
		gs = append(gs, g)
	}
	//查询所有的消费组的信息
	groupInfos, err := p.admin.DescribeConsumerGroups(gs)
	if err != nil {
		log.Printf("failed to list consumer groups info, err info: %v", err.Error())
		return groupIds, nil
	}
	//循环消费组的信息，获取分配的信息
	groupDict := make(map[string]any, 0)
	for _, groupInfo := range groupInfos {
		for _, member := range groupInfo.Members {
			//消费组内的partition分配情况
			assignment, err := member.GetMemberAssignment()
			if err != nil {
				log.Printf("failed to GetMemberAssignment info, err info: %v", err.Error())
				continue
			}
			//如果是分配到的topic刚好是指定的，那么就返回这个groupID
			for memberTopic, _ := range assignment.Topics {
				if memberTopic != topic {
					continue
				}
				if _, ok := groupDict[groupInfo.GroupId]; !ok {
					groupDict[groupInfo.GroupId] = struct{}{}
					groupIds = append(groupIds, groupInfo.GroupId)
				}
			}
		}
	}
	log.Printf("topic %s has groupIds: %v\n", topic, groupIds)
	return groupIds, nil
}

func (p *syncProducer) Send(topic string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(nil),
		Value: sarama.ByteEncoder(value),
	}
	fmt.Printf("send topic %s, msg %s\n", topic, string(value))
	_, _, err := p.producer.SendMessage(msg)
	return err
}

func (p *syncProducer) SendWithKey(topic string, key []byte, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	fmt.Printf("send topic %s, key %s, msg %s\n", topic, string(key), string(value))
	_, _, err := p.producer.SendMessage(msg)
	return err
}

func (p *syncProducer) RetrySend(ctx context.Context, topic string, messageBody []byte) error {
	err := retry.Do(
		func() error {
			return p.Send(topic, messageBody)
		},
		retry.Attempts(5),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			if n > 0 {
				log.Printf("failed to send topic %s, msg: %v, err: %v, retry %d times ...", topic, string(messageBody), err, n)
			}
		}),
		retry.RetryIf(func(err error) bool { return err != nil }),
		retry.MaxDelay(1*time.Second),
		retry.Context(ctx),
		retry.LastErrorOnly(true),
	)
	return err
}

func (p *syncProducer) Close() error {
	return p.producer.Close()
}
