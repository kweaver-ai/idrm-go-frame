package kafkax

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"golang.org/x/sync/errgroup"
	"log"
	"sync"
)

type consumerService struct {
	conf    *ConsumerConfig
	handles []MessageHandleDef

	ctx             context.Context
	mtx             sync.Mutex
	cancel          context.CancelFunc
	eg              *errgroup.Group
	consumerClients []sarama.ConsumerGroup
}

type MessageHandleDef struct {
	Topic   []string
	Handle  MsgHandleFunc
	Options []Option
}

func NewConsumerService(conf *ConsumerConfig) Consumer {
	return &consumerService{
		conf: conf,
	}
}

func warpForLog(cmd MsgHandleFunc) MsgHandleFunc {
	return func(ctx context.Context, msg *Message) bool {
		fmt.Printf("consumer topic %s, msg: %v\n", msg.Topic, string(msg.Value))
		return cmd(ctx, msg)
	}
}

func (s *consumerService) RegisterHandles(handle MsgHandleFunc, topics ...string) {
	mh := MessageHandleDef{
		Topic:   topics,
		Handle:  warpForLog(handle),
		Options: defaultOptions(),
	}
	s.handles = append(s.handles, mh)
}

func (s *consumerService) Subscribe(ctx context.Context) {
	go func() {
		if err := s.Start(ctx); err != nil {
			log.Printf("failed to subscribe msg err info: %s", err.Error())
			return
		}
	}()
}

func (s *consumerService) Start(ctx context.Context) error {
	s.init(ctx)

	s.checkHandles()

	if err := s.registerHandles(s.ctx); err != nil {
		return err
	}

	if err := s.eg.Wait(); err != nil {
		return err
	}

	for _, client := range s.consumerClients {
		if err := client.Close(); err != nil {
			log.Printf("failed to close consumer client, err: %v", err)
		}
	}

	return nil
}

func (s *consumerService) init(ctx context.Context) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	ctx, s.cancel = context.WithCancel(ctx)
	s.eg, s.ctx = errgroup.WithContext(ctx)
}

func (s *consumerService) Stop(_ context.Context) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	return nil
}

func (s *consumerService) checkHandles() {
	handles := s.getHandles()
	if len(handles) < 1 {
		panic("message handle def is empty")
	}

	set := map[string]struct{}{}
	for _, handle := range handles {
		if len(handle.Topic) < 1 || handle.Handle == nil {
			panic("incorrect definition of message handle")
		}

		for _, topic := range handle.Topic {
			if len(topic) < 1 {
				panic("incorrect definition of message handle, topic is empty")
			}

			if _, ok := set[topic]; ok {
				panic("incorrect definition of message handle, multiple handles for a single topic, topic: " + topic)
			}

			set[topic] = struct{}{}
		}
	}
}

func (s *consumerService) getHandles() []MessageHandleDef {
	return s.handles
}

func (s *consumerService) registerHandles(ctx context.Context) error {
	for _, handle := range s.handles {
		if err := s.registerHandle(ctx, handle); err != nil {
			return err
		}
	}

	return nil
}

func (s *consumerService) registerHandle(ctx context.Context, handle MessageHandleDef) error {
	client, optCfg, err := s.getClient(handle.Options)
	if err != nil {
		return err
	}

	s.consumerClients = append(s.consumerClients, client)
	s.eg.Go(func() error {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, handle.Topic, otelsarama.WrapConsumerGroupHandler(newKafkaConsumer(handle.Handle, optCfg))); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}

			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return nil
			}
		}
	})

	if optCfg.errHandle != nil {
		s.eg.Go(func() error {
			for {
				select {
				case err := <-client.Errors():
					optCfg.errHandle(err)

				case <-s.ctx.Done():
					return nil
				}
			}
		})
	}

	return nil
}

func (s *consumerService) getClient(options []Option) (sarama.ConsumerGroup, *optionConfig, error) {
	kafkaVersion, err := sarama.ParseKafkaVersion(s.conf.Version)
	if err != nil {
		return nil, nil, err
	}
	config := sarama.NewConfig()
	config.Version = kafkaVersion
	config.ClientID = s.conf.ClientID
	if s.conf.Mechanism != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLMechanism(s.conf.Mechanism)
		config.Net.SASL.Version = sarama.SASLHandshakeV1
		config.Net.SASL.Handshake = true
		config.Net.SASL.User = s.conf.UserName
		config.Net.SASL.Password = s.conf.Password
	}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	optCfg := newOptionConfig(*config)
	optCfg.trace = s.conf.Trace
	for _, option := range options {
		option.apply(optCfg)
	}

	client, err := sarama.NewConsumerGroup([]string{s.conf.Addr}, s.conf.GroupID, &optCfg.Config)
	if err != nil {
		log.Printf("failed to create kafka consumer client, err: %v", err)
		return nil, nil, err
	}

	return client, optCfg, nil
}
