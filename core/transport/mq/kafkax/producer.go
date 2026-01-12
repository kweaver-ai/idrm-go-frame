package kafkax

import "context"

type ProducerConfig struct {
	Addr      string
	UserName  string
	Password  string
	Mechanism string
	RecSize   int
}

type Producer interface {
	Send(topic string, value []byte) error
	SendWithKey(topic string, key []byte, value []byte) error
	RetrySend(ctx context.Context, topic string, messageBody []byte) error
	ListTopicGroupIds(topic string) ([]string, error)
	Close() error
}
