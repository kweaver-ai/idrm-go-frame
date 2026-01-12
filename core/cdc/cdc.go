package cdc

import (
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/options"
	"github.com/kweaver-ai/idrm-go-frame/core/store/redis"

	_ "github.com/go-sql-driver/mysql"
)

// SourceConf 生产者服务配置
type SourceConf struct {
	// DBOptions     *options.DBOptions
	Broker        string          // kafka 实例地址
	KafkaUser     string          // 用户名
	KafkaPassword string          // 密码
	Mechanism     string          // SASL的认知机制，一般是铭文PLAIN
	Version       string          // kafka的版本
	ClientID      string          // 客户端id
	RedisConfig   redis.RedisConf // redis配置，为了初始化分布式锁
	Sources       Sources
}

type Sources struct {
	Options options.DBOptions
	Source  []*CronConf `json:"source"` // 数据源、定时任务配置
}

type CronConf struct {
	Expression string `json:"expression"` // corn表达式

	Table               string `json:"table"`                 // 表名称
	Column              string `json:"column"`                // 字段名称
	IDColumnName        string `json:"id_column_name"`        // id字段名称
	TimestampColumnName string `json:"timestamp_column_name"` // 时间字段名称
}

// SyncMessage kafka message的payload
type SyncMessage struct {
	GroupIds  []string `json:"group_ids"` // 需要消费这条消息的消费者组
	Data      any      `json:"data"`      // 变更后的数据
	DB        string   `json:"db"`        // 数据库名称
	Schema    string   `json:"schema"`    // schema
	Table     string   `json:"table"`     // 表名称
	Operation string   `json:"op"`        // 变更类型 u更新 c新建
	TimeStamp int64    `json:"ts_ms"`     // 变更时间 毫秒时间戳
}

// Task 同步任务表结构体
type Task struct {
	Database  string    `json:"database"`
	Table     string    `json:"table"`
	Columns   string    `json:"columns"`
	Topic     string    `json:"topic"`
	GroupId   string    `json:"group_id"`
	Id        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SinkConf 消费者服务配置
type SinkConf struct {
	Broker        string // kafka 实例地址
	KafkaUser     string // 用户名
	KafkaPassword string // 密码
	Mechanism     string // SASL的认知机制，一般是铭文PLAIN
	ConsumerGroup string // 指定消费者组id
	ClientID      string // 客户端ID
	Version       string // kafka version
}
