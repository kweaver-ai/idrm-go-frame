package kafkax

import (
	"context"

	"github.com/kweaver-ai/idrm-go-frame/core/logx/zapx"

	"github.com/Shopify/sarama"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type kafkaConsumer struct {
	setup      func(sarama.ConsumerGroupSession) error
	cleanup    func(sarama.ConsumerGroupSession) error
	handle     func(context.Context, *sarama.ConsumerMessage) bool
	autoCommit bool
}

func newKafkaConsumer(handle MsgHandleFunc, cfg *optionConfig) sarama.ConsumerGroupHandler {
	return &kafkaConsumer{
		setup:   cfg.setup,
		cleanup: cfg.cleanup,
		handle: func(ctx context.Context, message *sarama.ConsumerMessage) bool {
			ctx = otel.GetTextMapPropagator().Extract(ctx, otelsarama.NewConsumerMessageCarrier(message))
			if cfg.trace != nil {
				ctx2, span := cfg.trace.Start(ctx, "consume message", trace.WithAttributes(
					semconv.MessagingOperationProcess,
				))
				ctx = ctx2
				defer span.End()
			}
			return handle(ctx, &Message{
				Topic:     message.Topic,
				Key:       string(message.Key),
				Value:     message.Value,
				Timestamp: message.Timestamp,
			})
		},
		autoCommit: cfg.Config.Consumer.Offsets.AutoCommit.Enable,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *kafkaConsumer) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	zapx.Infof("start recv msg from topic: %v", session.Claims())
	return c.setup(session)
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *kafkaConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	zapx.Infof("end recv msg from topic: %v", session.Claims())
	return c.cleanup(session)
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message := <-claim.Messages():
			if c.handle(session.Context(), message) {
				session.MarkMessage(message, "")
				if !c.autoCommit {
					session.Commit()
				}
			}

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}
