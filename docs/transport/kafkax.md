# 消息队列kafka使用

## 引入

建议在项目的 `infrastructure/mq/kafka`位置引入消息队列

### 消费者
```
func NewConsumer() kafkax.Consumer {
	Consumer := kafkax.NewConsumerService(&kafkax.ConsumerConfig{
		Version:   settings.ConfigInstance.Config.KafkaConf.Version,
		Addr:      settings.ConfigInstance.Config.KafkaConf.URI,
		ClientID:  settings.ConfigInstance.Config.KafkaConf.ClientId,
		UserName:  settings.ConfigInstance.Config.KafkaConf.Username,
		Password:  settings.ConfigInstance.Config.KafkaConf.Password,
		GroupID:   settings.ConfigInstance.Config.KafkaConf.GroupId,
		Mechanism: settings.ConfigInstance.Config.KafkaConf.Mechanism,
		Trace:     ar_trace.Tracer,
	})
	return Consumer
}
```

### 生产者
```
func NewSyncProducer() (kafkax.Producer, error) {
	producer, err := kafkax.NewSyncProducer(&kafkax.ProducerConfig{
		Addr:      settings.ConfigInstance.Config.KafkaConf.URI,
		UserName:  settings.ConfigInstance.Config.KafkaConf.Username,
		Password:  settings.ConfigInstance.Config.KafkaConf.Password,
		Mechanism: settings.ConfigInstance.Config.KafkaConf.Mechanism,
	})
	return producer, err
}
```

