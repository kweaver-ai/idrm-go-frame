package corn

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kweaver-ai/idrm-go-frame/core/options"
	"github.com/kweaver-ai/idrm-go-frame/core/store/redis"

	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/robfig/cron"
	"github.com/segmentio/kafka-go"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

const topicTemplate = "idrm.cdc.%s.%s"

type SyncMiddleware struct {
	once     sync.Once
	producer map[string]*kafka.Writer
	consumer *kafka.Reader
	db       *sql.DB
	cron     *cron.Cron
	rs       *redsync.Redsync

	cornMap map[string]*CornConf
}

type SyncConf struct {
	DBOptions   *options.DBOptions
	Brokers     []string
	RedisConfig redis.RedisConf
	Sources     []*CornConf
}

type SyncMessage struct {
	Data      any    `json:"data"` // 变更后的数据
	Source    source `json:"source"`
	Operation string `json:"op"`    // 变更类型 u更新 c新建 d删除
	TimeStamp int64  `json:"ts_ms"` // 变更时间 毫秒时间戳
}

type source struct {
	DB     string `json:"db"`     // 数据库名称
	Schema string `json:"schema"` // schema
	Table  string `json:"table"`  // 表名称
}

type CornConf struct {
	Expression string `json:"expression"` // corn表达式

	Type                string `json:"type"`                  // 数据库类型
	Host                string `json:"host"`                  // 数据库地址
	Username            string `json:"username"`              // 用户名
	Password            string `json:"password"`              // 密码
	DB                  string `json:"db"`                    // 库名称
	Schema              string `json:"schema"`                // schema
	Table               string `json:"table"`                 // 表名称
	Column              string `json:"column"`                // 字段名称
	IDColumnName        string `json:"id_column_name"`        // id字段名称
	TimestampColumnName string `json:"timestamp_column_name"` // 时间字段名称
}

func NewSyncMiddleware(db *sql.DB, conf SyncConf) (*SyncMiddleware, error) {
	//if err := checkConf(conf); err != nil {
	// return nil, err
	//}
	middle := &SyncMiddleware{}

	middle.once.Do(func() {
		writerMap := make(map[string]*kafka.Writer, len(conf.Sources))
		for _, source := range conf.Sources {
			database := source.DB
			table := source.Table
			topic := fmt.Sprintf(topicTemplate, database, table)
			if _, ok := writerMap[topic]; !ok {
				writer := &kafka.Writer{
					Addr:  kafka.TCP(conf.Brokers...),
					Topic: topic,
				}
				writerMap[topic] = writer
			}
		}
		middle.producer = writerMap
	})

	db, err := sql.Open("mysql", conf.DBOptions.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}
	if err := initDB(db); err != nil {
		return nil, fmt.Errorf("failed to init database: %v", err)
	}

	client := goredislib.NewClient(&goredislib.Options{
		Addr:     conf.RedisConfig.Host,
		Password: conf.RedisConfig.Pass,
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	middle.rs = redsync.New(pool)
	middle.cron = cron.New()

	middle.cornMap = loadCornMap(conf.Sources)

	return middle, nil
}

func initDB(db *sql.DB) error {
	// 在接入中间件的服务数据库里创建同步记录表
	return nil
}

func loadCornMap(corns []*CornConf) map[string]*CornConf {
	cornMap := make(map[string]*CornConf)
	for _, corn := range corns {
		// 同一表字段的corn表达式唯一
		key := corn.DB + corn.Table + corn.Column + corn.Expression
		if _, ok := cornMap[key]; !ok {
			cornMap[key] = corn
		}
	}
	return cornMap
}

func (s *SyncMiddleware) Close() {
	for _, writer := range s.producer {
		writer.Close()
	}
	s.consumer.Close()
	s.db.Close()
	s.cron.Stop()
}

func (s *SyncMiddleware) ConsumeMessages(ctx context.Context) {
	for {
		msg, err := s.consumer.FetchMessage(ctx)
		if err != nil {
			log.Printf("Failed to consume message: %v", err)
			continue
		}
		fmt.Printf("Received message: key=%s, value=%s\n", string(msg.Key), string(msg.Value))

		// 在这里处理接收到的消息，可以将消息写入数据库或执行其他操作
		err = s.consumer.CommitMessages(ctx, msg)
		if err != nil {
			return
		}
	}
}

func (s *SyncMiddleware) ScheduleIncrementalSync() error {
	for _, v := range s.cornMap {
		c := v
		go func() {
			err := s.cron.AddFunc(c.Expression, func() {
				// 在这里执行增量同步的逻辑，可以从数据库中获取上次同步的记录，并根据记录进行增量同步
				mutex := s.rs.NewMutex(c.Table+c.Column, redsync.WithValue("incr"))
				mutex.TryLock()
				defer mutex.Unlock()

				if query, err := s.db.Query(""); err != nil {
					return
				} else {
					for query.Next() {
						syncMsg := SyncMessage{
							Data: nil,
							Source: source{
								DB:     c.DB,
								Schema: c.Schema,
								Table:  c.Table,
							},
							Operation: "c",
							TimeStamp: time.Now().UnixMilli(),
						}
						bytes, _ := json.Marshal(syncMsg)
						topic := fmt.Sprintf(topicTemplate, c.DB, c.Table)
						msg := kafka.Message{
							Topic: topic,
							Key:   []byte(c.DB + c.Table + c.Column),
							Value: bytes,
							Time:  time.Now(),
						}
						s.dispatchMessage(context.Background(), topic, msg)
					}
				}

			})
			if err != nil {
				log.Printf("failed to schedule incremental sync: %v", err)
				return
			}
		}()
	}

	return nil
}

func (s *SyncMiddleware) dispatchMessage(ctx context.Context, topic string, msg kafka.Message) error {
	return s.producer[topic].WriteMessages(ctx, msg)
}

func (s *SyncMiddleware) FullSync() error {
	return nil
}

// ListenAndConsume 消费端调用
func (s *SyncMiddleware) ListenAndConsume() error {
	// 在这个里面消费消息，写到数据库
	return nil
}
