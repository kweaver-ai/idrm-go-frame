package cdc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/store/gormx"
	"github.com/kweaver-ai/idrm-go-frame/core/transport/mq/kafkax"

	"github.com/pkg/errors"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/uuid"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
)

// 第一个占位符是库名，第二个占位符是表名
const topicTemplate = `idrm.cdc.%s.%s`

const (
	mysqlTaskTableSql = "CREATE TABLE IF NOT EXISTS `cdc_task`\n(\n `database` VARCHAR(255) NOT NULL COMMENT '同步库名',\n `table` VARCHAR(255) NOT NULL COMMENT '同步表名',\n `columns` VARCHAR(255) NOT NULL COMMENT '同步的列，多个列写在一起，用 , 隔开',\n `topic` VARCHAR(255) NOT NULL COMMENT '数据变动投递消息的topic',\n `group_id` VARCHAR(255) NOT NULL COMMENT '当前记录对应的group id',\n `id` VARCHAR(255) NOT NULL COMMENT '当前同步记录id',\n `updated_at` DATETIME(3) NOT NULL COMMENT '当前同步记录时间'\n) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4  COLLATE=utf8mb4_unicode_ci;"
	dm8TaskTableSql   = `CREATE TABLE IF NOT EXISTS cdc_task ( "database" VARCHAR(255) NOT NULL,  "table" VARCHAR(255) NOT NULL, "columns" VARCHAR(255) NOT NULL, "topic" VARCHAR(255) NOT NULL , "group_id" VARCHAR(255) NOT NULL, "id" VARCHAR(255) NOT NULL, "updated_at" DATETIME(3) NOT NULL) ;`
)

var sourceOnce sync.Once

type Source struct {
	cronWg sync.WaitGroup

	producer   kafkax.Producer
	db         *sql.DB
	dbType     string
	dbDatabase string
	dbSchema   string

	cron *cron.Cron
	rs   *redsync.Redsync

	cronMap map[string]*CronConf
}

func InitSource(config *SourceConf) (s *Source, err error) {
	source := &Source{
		cronWg:     sync.WaitGroup{},
		cron:       cron.New(),
		dbType:     config.Sources.Options.DBType,
		dbDatabase: config.Sources.Options.Database,
		dbSchema:   "",
	}
	//初始化MQ
	publisher, err := kafkax.NewSyncProducer(
		&kafkax.ProducerConfig{
			Addr:      config.Broker,
			UserName:  config.KafkaUser,
			Password:  config.KafkaPassword,
			Mechanism: config.Mechanism,
		},
	)
	if err != nil {
		log.Printf("failed to create publisher, err info: %v", err.Error())
		return
	}
	source.producer = publisher
	//初始化定时任务
	if err = source.initCronMap(config); err != nil {
		return
	}
	//初始化redis
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     config.RedisConfig.Host,
		Password: config.RedisConfig.Pass,
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	source.rs = redsync.New(pool)
	//初始化数据库
	gormDB, err := config.Sources.Options.NewClient()
	if err != nil {
		log.Printf("failed to open db, err info: %v", err.Error())
		return source, err
	}
	source.db, err = gormDB.DB()
	if err != nil {
		log.Printf("failed to start db, err info: %v", err.Error())
		return source, err
	}
	return source, nil
}

func (s *Source) initCronMap(config *SourceConf) error {
	cronMap := make(map[string]*CronConf)
	for _, conf := range config.Sources.Source {
		if len(conf.Table) == 0 || len(conf.Column) == 0 {
			return errors.New("init cron map failed, empty table name or column.")
		}
		key := fmt.Sprintf(topicTemplate, config.Sources.Options.Database, conf.Table)
		if _, ok := cronMap[key]; ok {
			return errors.New("init cron map failed, duplicated table name.")
		}
		cronMap[key] = conf
	}
	s.cronMap = cronMap
	return nil
}

func (s *Source) Start() {
	err := s.createCdcTaskTable()
	if err != nil {
		log.Printf("create cdc_task table error, [%s]\n", err.Error())
		return
	}

	log.Printf("create cdc_task table finished\n")

	for topic := range s.cronMap {
		// 发送一条空消息到topic里，让topic有正常的初始offset，否则会是 kafkac.OffsetInvalid，监测不到消费者组
		s.topicOffsetInit(topic)

		log.Printf("listen topic [%s] started...\n", topic)
		t := topic
		table := s.cronMap[topic].Table
		column := s.cronMap[topic].Column
		idColumnName := s.cronMap[topic].IDColumnName
		timestampColumnName := s.cronMap[topic].TimestampColumnName
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			for {
				<-ticker.C
				// 监听 group id 数量
				newGroupIds := s.listenTopicGroupIds(t)
				if len(newGroupIds) > 0 {
					s.doFullSync(table, column, idColumnName, timestampColumnName, t, newGroupIds)
				}
			}
		}()
	}
	// go s.listenTopicGroupIds()
	go s.dispatchCron()
	select {}
}

// doFullSync 发起全量同步
func (s *Source) doFullSync(table, column, idColumnName, timestampColumnName, topic string, newGroupIds []string) {
	// 将新增的topic写入cdc_task，发起全量同步
	mutex := s.rs.NewMutex(topic)
	var lockErr error
	lockErr = mutex.Lock()
	for lockErr != nil {
		log.Printf("mutex.Lock topic [%s] failed, err info: %v", topic, lockErr.Error())
		time.Sleep(5 * time.Second)
		// 自旋获取锁，阻塞
		lockErr = mutex.Lock()
	}

	rowData := make([]map[string]interface{}, 0)
	newColumn := column
	if !s.inArray(strings.Split(newColumn, ","), idColumnName) {
		newColumn = newColumn + ", " + idColumnName
	}
	if !s.inArray(strings.Split(newColumn, ","), timestampColumnName) {
		newColumn = newColumn + ", " + timestampColumnName
	}
	query := "SELECT %s FROM %s;"
	stmt, err := s.db.Prepare(fmt.Sprintf(query, newColumn, table))
	if err != nil {
		log.Printf("failed to open stmt: %s, err info: %v", fmt.Sprintf(query, newColumn, table), err.Error())
		return
	}
	defer stmt.Close()
	source, err := stmt.Query()
	if err != nil {
		log.Printf("failed to query cron cdc_task, topic: %s, table: %s, err info: %s", topic, table, err.Error())
		return
	}
	defer source.Close()

	columns, _ := source.Columns()
	columnTypes, _ := source.ColumnTypes()

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	for source.Next() {
		if err := source.Scan(values...); err != nil {
			log.Fatal(err)
		}
		row := make(map[string]interface{})
		for i, col := range values {
			colValue := reflect.ValueOf(col).Elem().Interface()
			colType := columnTypes[i].DatabaseTypeName()
			//fmt.Printf("columns[%v]: %v, colType: %s, colValue: %s\n", i, columns[i], colType, colValue)
			if !strings.Contains(newColumn, columns[i]) {
				continue
			}
			switch colType {
			case "INT", "BIGINT":
				if v, ok := colValue.(int64); ok {
					row[columns[i]] = v
				}
			case "VARCHAR", "CHAR", "TEXT":
				if v, ok := colValue.([]uint8); ok {
					row[columns[i]] = string(v)
				}
			case "DATETIME", "DATE":
				if v, ok := colValue.(time.Time); ok {
					row[columns[i]] = v
				}
			default:
				row[columns[i]] = colValue
			}
		}
		rowData = append(rowData, row)
	}

	var latestRecord map[string]interface{}
	if len(rowData) > 0 {
		latestRecord = rowData[len(rowData)-1]
	} else {
		// 没有查到数据 退出
		return
	}
	var id string
	var updatedAt time.Time
	if v, ok := latestRecord[idColumnName]; ok {
		switch v.(type) {
		case int64:
			id = strconv.FormatInt(v.(int64), 10)
		case time.Time:
			id = v.(time.Time).Format("2006-01-02 15:04:05.999")
		}
	}
	if v, ok := latestRecord[timestampColumnName]; ok {
		updatedAt = v.(time.Time)
	}
	// s.updateTask(topic, column, id , updatedAt)
	tasks := make([]*Task, 0, len(newGroupIds))
	for _, groupId := range newGroupIds {
		task := &Task{
			Database:  s.dbDatabase,
			Table:     table,
			Columns:   newColumn,
			Topic:     topic,
			GroupId:   groupId,
			Id:        id,
			UpdatedAt: updatedAt,
		}
		tasks = append(tasks, task)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	if err := s.createTask(tx, tasks); err != nil {
		taskStr, _ := json.Marshal(tasks)
		log.Printf("failed to createTask,tasks: %s, err info: %s\n", string(taskStr), err.Error())
		tx.Rollback()
		return
	}

	syncMessages := make([]*SyncMessage, len(rowData), len(rowData))
	for i, data := range rowData {
		syncMessages[i] = &SyncMessage{
			GroupIds:  newGroupIds,
			Data:      data,
			DB:        s.dbDatabase,
			Schema:    s.dbSchema,
			Table:     table,
			Operation: "c",
			TimeStamp: time.Now().UnixMilli(),
		}
	}
	if err := s.packAndPublishMsg(topic, syncMessages...); err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	mutex.Unlock()
}

// createCdcTaskTable 创建 cdc_task 表
func (s *Source) createCdcTaskTable() (err error) {
	if strings.Contains(gormx.DriveDm, s.dbType) {
		_, err = s.db.Exec(dm8TaskTableSql)
	} else {
		_, err = s.db.Exec(mysqlTaskTableSql)
	}
	return
}

// createTask 定时任务表创建记录
func (s *Source) createTask(tx *sql.Tx, tasks []*Task) error {
	query := "INSERT INTO cdc_task VALUES(?, ?, ?, ?, ?, ?, ?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, task := range tasks {
		_, err := stmt.Exec(task.Database, task.Table, task.Columns, task.Topic, task.GroupId, task.Id, task.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// listenTopicGroupIds 监听指定 topic 下的 group id，如果有新增的，放到切片中返回
func (s *Source) listenTopicGroupIds(topic string) []string {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recover from err: %v", err)
		}
	}()
	newGroupIds := make([]string, 0)

	groupIds, err := s.producer.ListTopicGroupIds(topic)
	if err != nil {
		log.Printf("publisher listen topic group ids failed, err info: %s", err.Error())
		return newGroupIds
	}

	// 查询数据库中对应topic的所有group id
	rows, err := s.db.Query("SELECT `group_id` FROM cdc_task WHERE topic = ?;", topic)
	if err != nil {
		log.Printf("publisher query topic group ids failed, topic: %s, err info: %s", "topic", err.Error())
		return newGroupIds
	}
	defer rows.Close()

	groupsInDB := make(map[string]struct{}, 0)
	for rows.Next() {
		var groupId string
		if err := rows.Scan(&groupId); err != nil {
			log.Printf("publisher query db topic group ids failed, topic: %s, err info: %s", topic, err.Error())
			continue
		}
		groupsInDB[groupId] = struct{}{}
	}

	if len(groupsInDB) < len(groupIds) {
		for _, group := range groupIds {
			if _, ok := groupsInDB[group]; !ok {
				newGroupIds = append(newGroupIds, group)
			}
		}
	}
	return newGroupIds
}

func (s *Source) topicOffsetInit(topic string) error {
	syncMessage := &SyncMessage{
		GroupIds:  []string{},
		Data:      "",
		DB:        "",
		Schema:    "",
		Table:     "",
		Operation: "",
		TimeStamp: time.Now().UnixMilli(),
	}
	if err := s.packAndPublishMsg(topic, syncMessage); err != nil {
		log.Printf("topicOffsetInit failed, topic: %s, err info: %s", topic, err.Error())
		return err
	}

	time.Sleep(5 * time.Second)

	return nil
}

// dispatchCron 分发定时任务（增量同步）
func (s *Source) dispatchCron() {
	s.cronWg.Add(len(s.cronMap))
	for t, job := range s.cronMap {
		topic := t
		c := job
		go func() {
			defer s.cronWg.Done()
			if err := s.cron.AddFunc(c.Expression, func() {
				// 查询数据库，发送消息
				// 查询同步任务表
				tasks, err := s.queryTask(topic)
				if err != nil {
					return
				}
				if len(tasks) == 0 {
					log.Printf("topic [%s] not found from cdc_task, waiting for init...", topic)
					return
				}
				id := tasks[0].Id
				updatedAt := tasks[0].UpdatedAt

				// 查询源数据表
				// 1.select * from source ... -> 多条记录
				query := "SELECT COUNT(*) FROM %s WHERE ((%s = ? AND %s > ?) OR %s > ?) AND %s < ? ORDER BY %s, %s ASC"
				sql := fmt.Sprintf(query, c.Table, c.TimestampColumnName, c.IDColumnName, c.TimestampColumnName, c.TimestampColumnName, c.TimestampColumnName, c.IDColumnName)
				countRows, err := s.db.Query(sql, updatedAt, id, updatedAt, time.Now())
				if err != nil {
					log.Printf("failed to query db, err info: %v\n", err.Error())
					return
				}
				defer countRows.Close()

				var count int64
				for countRows.Next() {
					countRows.Scan(&count)
				}
				// 比较是否需要执行增量同步任务
				if count == 0 {
					// 无需执行，直接返回
					log.Printf("topic [%s] is already updated to id: %v, updatedAt: %v, no need to sync", topic, id, updatedAt)
					return
				}

				// 需要执行
				mutex := s.rs.NewMutex(topic)
				if lockErr := mutex.Lock(); lockErr != nil {
					// 增量同步获取锁失败，两种可能
					// 1.增量同步进行中，本次增量冲突，放弃。
					// 2.全量同步进行中，阻塞当前增量同步，也直接放弃，等待下一次增量同步。
					log.Printf("mutex.Lock topic [%s] failed err info: %v", topic, err.Error())
					return
				}
				defer mutex.Unlock()

				rowData := s.querySourceTable(c.Table, c.Column, c.IDColumnName, c.TimestampColumnName, topic, id, updatedAt)
				//log.Printf("row data: %d\n", len(rowData))

				// cdc_task 表中所有的 group id
				groupIds := make([]string, 0, len(tasks))
				for _, task := range tasks {
					groupIds = append(groupIds, task.GroupId)
				}
				log.Printf("topic: [%s], groupIds: %s, columns: [%s] need to do incr sync\n", topic, groupIds, c.Column)

				// 更新task表，发送消息到队列
				syncMessages := make([]*SyncMessage, len(rowData), len(rowData))
				for i, data := range rowData {
					syncMessages[i] = &SyncMessage{
						GroupIds:  groupIds,
						Data:      data,
						DB:        s.dbDatabase,
						Schema:    s.dbSchema,
						Table:     c.Table,
						Operation: "u", // TODO 判断操作类型c u d
						TimeStamp: time.Now().UnixMilli(),
					}
				}

				tx, err := s.db.Begin()
				if err != nil {
					log.Println("failed to begin tx, err info: ", err.Error())
					return
				}

				latestRecord := rowData[len(rowData)-1]

				var latestId string
				switch latestRecord[c.IDColumnName].(type) {
				case int64:
					latestId = strconv.FormatInt(latestRecord[c.IDColumnName].(int64), 10)
				case time.Time:
					latestId = latestRecord[c.IDColumnName].(time.Time).Format("2006-01-02 15:04:05.999")
				}

				latestUpdated := latestRecord[c.TimestampColumnName].(time.Time)

				// 这里要更新的updatedAt应该是从row data里面查出来的最大的记录
				if err := s.updateTask(tx, topic, groupIds, latestId, latestUpdated); err != nil {
					log.Printf("failed to update task, topic [%s], err info: %s\n", topic, err.Error())
					tx.Rollback()
					return
				}

				if err := s.packAndPublishMsg(topic, syncMessages...); err != nil {
					log.Printf("s.packAndProduceMsg failed, topic [%s], err info: %s", topic, err.Error())
					tx.Rollback()
					return
				}
				tx.Commit()
				log.Printf("topic: [%s], groupIds: %s, columns [%s] do incr sync finished, updated  to id %s, updated_at: %s\n", topic, groupIds, c.Column, id, updatedAt)

			}); err != nil {
				log.Printf("failed to add cron, err info: %s", err.Error())
				return
			}
		}()
	}
	s.cronWg.Wait()
	s.cron.Start()
}

// querySourceTable 查询源数据表
func (s *Source) querySourceTable(table, column, idColumnName, timestampColumnName, topic, id string, updatedAt time.Time) (rowData []map[string]interface{}) {
	rowData = make([]map[string]interface{}, 0)

	// 确保查询的列中包括 id updated_at
	newColumn := column
	if !s.inArray(strings.Split(newColumn, ","), idColumnName) {
		newColumn = newColumn + ", " + idColumnName
	}
	if !s.inArray(strings.Split(newColumn, ","), timestampColumnName) {
		newColumn = newColumn + ", " + timestampColumnName
	}

	query := "SELECT %s FROM %s WHERE ((%s = ? AND %s > ?) OR %s > ?) AND %s < ? ORDER BY %s, %s ASC"
	sql := fmt.Sprintf(query, newColumn, table, timestampColumnName, idColumnName, timestampColumnName, timestampColumnName, timestampColumnName, idColumnName)
	source, err := s.db.Query(sql, updatedAt, id, updatedAt, time.Now())
	if err != nil {
		log.Printf("failed to query cron cdc_task, topic: %s, err info: %s", topic, err.Error())
		return
	}
	defer source.Close()

	// 2.根据配置字段动态映射结构体
	columns, _ := source.Columns()
	columnTypes, _ := source.ColumnTypes()

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	for source.Next() {
		if err := source.Scan(values...); err != nil {
			log.Fatal(err)
		}
		row := make(map[string]interface{})
		for i, col := range values {
			colValue := reflect.ValueOf(col).Elem().Interface()
			colType := columnTypes[i].DatabaseTypeName()
			//fmt.Printf("columns[%v]: %v, colType: %s, colValue: %s\n", i, columns[i], colType, colValue)
			if !strings.Contains(newColumn, columns[i]) {
				continue
			}
			switch colType {
			case "INT", "BIGINT":
				if v, ok := colValue.(int64); ok {
					row[columns[i]] = v
				}
			case "VARCHAR", "CHAR", "TEXT":
				if v, ok := colValue.([]uint8); ok {
					row[columns[i]] = string(v)
				}
			case "DATETIME", "DATE":
				if v, ok := colValue.(time.Time); ok {
					row[columns[i]] = v
				}
			default:
				row[columns[i]] = colValue
			}
		}
		rowData = append(rowData, row)
	}
	return
}

// queryTask 查询任务表
func (s *Source) queryTask(topic string) (tasks []*Task, err error) {
	tasks = make([]*Task, 0)
	rows, err := s.db.Query("SELECT * FROM cdc_task WHERE topic = ? ORDER BY id desc, updated_at desc limit 1;", topic)
	if err != nil {
		log.Printf("failed to query cron cdc_task, topic: %s, err info: %s", topic, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task = &Task{}
		if err = rows.Scan(&task.Database, &task.Table, &task.Columns, &task.Topic, &task.GroupId, &task.Id, &task.UpdatedAt); err != nil {
			log.Printf("failed to scan result to task, err info: %s\n", err.Error())
			return
		}
		log.Printf("task query result: %s", task)
		tasks = append(tasks, task)
	}
	return
}

func (s *Source) updateTask(tx *sql.Tx, topic string, groupIds []string, id string, updatedAt time.Time) error {
	// database/sql 不支持预编译 in ?，需要做一次拼接转换
	placeholders := strings.Repeat("?, ", len(groupIds)-1) + "?"

	sql := "UPDATE cdc_task SET id = ?, updated_at = ? WHERE topic = ? AND group_id IN (%s);"
	stmt, err := tx.Prepare(fmt.Sprintf(sql, placeholders))
	if err != nil {
		return err
	}
	defer stmt.Close()

	// 转化参数
	args := []interface{}{id, updatedAt, topic}
	for _, v := range groupIds {
		args = append(args, v)
	}
	log.Println("stmt.Exec args: ", args)

	if _, err = stmt.Exec(args...); err != nil {
		return err
	}

	return nil
}

func (s *Source) packAndPublishMsg(topic string, messages ...*SyncMessage) error {
	for _, message := range messages {
		bytes, _ := json.Marshal(message)
		if err := s.producer.SendWithKey(topic, []byte(uuid.NewString()), bytes); err != nil {
			log.Printf("producer msg error %v", err.Error())
			return err
		}
	}
	return nil
}

func (s *Source) inArray(array []string, x string) bool {
	for _, item := range array {
		if item == x {
			return true
		}
	}

	return false
}

func (s *Source) Stop() {
	s.db.Close()
	s.cron.Stop()
	s.producer.Close()
}
