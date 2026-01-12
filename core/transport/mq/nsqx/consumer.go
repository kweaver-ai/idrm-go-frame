package nsqx

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
	"net/http"
)

// Consumer MQ客户端接口
type Consumer interface {
	Register(topic string, cmd nsq.HandlerFunc)
	Close()
}

type TopicHandler struct {
	consumer *nsq.Consumer
	cmd      []nsq.HandlerFunc
}

type ConsumerService struct {
	config    Config
	NSQConfig *nsq.Config
	cs        map[string]*TopicHandler
}

func NewNSQConsumer(c Config) *ConsumerService {
	return &ConsumerService{
		config:    c,
		NSQConfig: nsq.NewConfig(),
		cs:        make(map[string]*TopicHandler),
	}
}

func warpForLog(topic string, cmd nsq.HandlerFunc) nsq.HandlerFunc {
	return func(msg *nsq.Message) error {
		fmt.Printf("consumer addr: %s, msg: %v\n", topic, string(msg.Body))
		return cmd(msg)
	}
}

// Register 注册handler 暂不支持多个handler消费同一个topic
func (c *ConsumerService) Register(topic string, cmd nsq.HandlerFunc) {
	if err := c.create(topic); err != nil {
		panic(err)
	}
	if err := c.subscribe(topic, c.config.Channel, warpForLog(topic, cmd)); err != nil {
		panic(err)
	}
}

func (c *ConsumerService) subscribe(topic string, channel string, cmd nsq.HandlerFunc) error {
	consumer, err := nsq.NewConsumer(topic, channel, c.NSQConfig) //创建消费者
	if err != nil {
		return fmt.Errorf("NewConsumer error:%v", err.Error())
	}
	// 定义 nsq 处理器
	consumer.AddHandler(cmd)

	// 连接 lookupd->nsqd
	err = consumer.ConnectToNSQLookupd(c.config.LookupdHost)
	if err != nil {
		return fmt.Errorf("consumer ConnectToNSQLookupd error: %v", err.Error())
	}
	c.addConsumer(topic, cmd, consumer)
	return nil
}

func (c *ConsumerService) create(topic string) error {
	addr := fmt.Sprintf("http://%s/topic/create?topic=%s", c.config.HttpHost, topic)

	request, err := http.NewRequest(http.MethodPost, addr, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create topic error %v")
	}
	return nil
}

func (c *ConsumerService) addConsumer(topic string, cmd nsq.HandlerFunc, consumer *nsq.Consumer) {
	th, ok := c.cs[topic]
	if !ok {
		c.cs[topic] = &TopicHandler{
			consumer: consumer,
			cmd:      []nsq.HandlerFunc{cmd},
		}
		return
	}
	th.consumer = consumer
	th.cmd = append(th.cmd, cmd)
	c.cs[topic] = th
}

func (c *ConsumerService) Close() {
	for topic, handler := range c.cs {
		fmt.Printf("consumer topic Close", zap.Any("topic", topic))
		handler.consumer.Stop()
	}
}
