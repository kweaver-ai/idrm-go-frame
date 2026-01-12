package gormx

import (
	"fmt"
	"sync"

	"github.com/kweaver-ai/idrm-go-frame/core/telemetry/log"

	"gorm.io/gorm"
)

var onceDict map[string]*sync.Once

func init() {
	onceDict = make(map[string]*sync.Once)
}

func getOnce(dbName string) *sync.Once {
	once, ok := onceDict[dbName]
	if !ok {
		once = &sync.Once{}
		onceDict[dbName] = once
	}
	return once
}

func ReleaseFunc(client *gorm.DB) func() {
	return func() {
		log.Info("closing the data resources")
	}
}

func NewOnce(e Options) (client *gorm.DB, err error) {
	once := getOnce(e.DBName)
	executed := false
	once.Do(func() {
		executed = true
		client, err = New(e)
	})
	if !executed {
		err = fmt.Errorf("db %v client has created", e.DBName)
	}
	return client, err
}
