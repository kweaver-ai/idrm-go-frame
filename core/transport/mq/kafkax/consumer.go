package kafkax

import (
	"context"
	"fmt"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/transport"

	"github.com/Shopify/sarama"
	"go.opentelemetry.io/otel/trace"
)

const SASLTypePlaintext = sarama.SASLTypePlaintext

type ConsumerConfig struct {
	Addr      string
	Channel   string
	ClientID  string
	GroupID   string
	UserName  string
	Password  string
	Mechanism string
	Version   string
	Trace     trace.Tracer
}
type Message struct {
	Topic     string
	Key       string
	Value     []byte
	Timestamp time.Time
}

type MsgHandleFunc func(ctx context.Context, msg *Message) bool

type Consumer interface {
	transport.Server                          //跟着gin框架一起启动
	RegisterHandles(MsgHandleFunc, ...string) //注册处理方法
	Subscribe(ctx context.Context)            //开始处理消息，手动开启
}

func Wrap(handler func(ctx context.Context, msg *Message) error) MsgHandleFunc {
	return func(ctx context.Context, msg *Message) bool {
		if err := handler(ctx, msg); err != nil {
			fmt.Printf("topic %s, value: %v, consumer error: %v", msg.Topic, string(msg.Value), err.Error())
			return false
		}
		return true
	}
}
