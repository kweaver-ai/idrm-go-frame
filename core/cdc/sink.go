package cdc

import (
	"context"

	"github.com/kweaver-ai/idrm-go-frame/core/transport/mq/kafkax"
)

type Sink struct {
	consumer kafkax.Consumer
}

func InitSink(config *SinkConf) *Sink {
	subscriber := kafkax.NewConsumerService(&kafkax.ConsumerConfig{
		Addr:      config.Broker,
		GroupID:   config.ConsumerGroup,
		ClientID:  config.ClientID,
		UserName:  config.KafkaUser,
		Password:  config.KafkaPassword,
		Mechanism: config.Mechanism,
		Version:   config.Version,
	})
	return &Sink{consumer: subscriber}
}

type MsgHandleFunc func(ctx context.Context, msg *kafkax.Message) error

func (s *Sink) RegisterLocal(topic string, handler MsgHandleFunc) {
	s.consumer.RegisterHandles(kafkax.Wrap(handler), topic)
}

func (s *Sink) Register(topic string, handler kafkax.MsgHandleFunc) {
	s.consumer.RegisterHandles(handler, topic)
}

func (s *Sink) Start(ctx context.Context) error {
	return s.consumer.Start(ctx)
}
