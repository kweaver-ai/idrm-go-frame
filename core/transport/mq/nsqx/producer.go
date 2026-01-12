package nsqx

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"log"
)

type Producer interface {
	Send(topic string, message []byte) error
}

type SyncProducer struct {
	producer *nsq.Producer
}

func NewSyncProducer(c Config) Producer {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(c.Host, config)
	if err != nil {
		log.Panicf("new SyncProducer error %v", err)
	}
	return &SyncProducer{producer: producer}
}

func (n SyncProducer) Send(topic string, message []byte) error {
	if err := n.producer.Publish(topic, message); err != nil {
		return fmt.Errorf("send message error %v", err)
	}
	return nil
}
