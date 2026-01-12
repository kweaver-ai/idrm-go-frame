package kafkax

import (
	"github.com/Shopify/sarama"
	"go.opentelemetry.io/otel/trace"
	"log"
)

type optionConfig struct {
	sarama.Config
	trace     trace.Tracer
	setup     func(sarama.ConsumerGroupSession) error
	cleanup   func(sarama.ConsumerGroupSession) error
	errHandle func(error)
}

func defaultOptions() []Option {
	return []Option{
		WithErrHandle(func(err error) {
			log.Printf("failed to consume kafka msg, err: %v", err)
		}),
		WithAutoCommit(false),
	}
}

func newOptionConfig(config sarama.Config) *optionConfig {
	return &optionConfig{
		Config: config,
		setup: func(session sarama.ConsumerGroupSession) error {
			return nil
		},
		cleanup: func(session sarama.ConsumerGroupSession) error {
			return nil
		},
		errHandle: func(err error) {
			log.Printf("kafka consumer err: %v", err)
		},
	}
}

type Option interface {
	apply(*optionConfig)
}

type OptionFunc func(*optionConfig)

func (f OptionFunc) apply(c *optionConfig) {
	f(c)
}

func WithSetup(f func(sarama.ConsumerGroupSession) error) Option {
	return OptionFunc(func(config *optionConfig) {
		if f != nil {
			config.setup = f
		}
	})
}

func WithCleanup(f func(sarama.ConsumerGroupSession) error) Option {
	return OptionFunc(func(config *optionConfig) {
		if f != nil {
			config.cleanup = f
		}
	})
}

func WithAutoCommit(auto bool) Option {
	return OptionFunc(func(config *optionConfig) {
		config.Consumer.Offsets.AutoCommit.Enable = auto
	})
}

func WithErrHandle(f func(error)) Option {
	return OptionFunc(func(config *optionConfig) {
		config.errHandle = f
	})
}
