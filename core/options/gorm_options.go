package options

import (
	"strings"

	"github.com/kweaver-ai/idrm-go-frame/core/store/gormx"

	"gorm.io/gorm"
)

// DBOptions defines options for  database.
type DBOptions struct {
	DBType                string `json:"dbtype,omitempty"                   mapstructure:"db-type"`
	Host                  string `json:"host,omitempty"                     mapstructure:"host"`
	Port                  string `json:"port,omitempty"                     mapstructure:"port"`
	Username              string `json:"username,omitempty"                 mapstructure:"username"`
	Password              string `json:"password"                           mapstructure:"password"`
	Database              string `json:"database"                           mapstructure:"database"`
	Config                string `json:"config" 							mapstructure:"config"`
	MaxIdleConnections    int    `json:"max-idle-connections,omitempty"     mapstructure:"max-idle-connections"`
	MaxOpenConnections    int    `json:"max-open-connections,omitempty"     mapstructure:"max-open-connections"`
	MaxConnectionIdleTime int    `json:"max-connection-idle-time"   		mapstructure:"max-connection-idle-time"` //秒
	MaxConnectionLifeTime int    `json:"max-connection-life-time" 			mapstructure:"max-connection-life-time"`  //秒
	LogLevel              int    `json:"log-level"                          mapstructure:"log-level"`
	IsDebug               bool   `json:"isdebug"                            mapstructure:"is-debug"`
	TablePrefix           string `json:"tableprefix"                        mapstructure:"table-prefix"`
}

// NewClient create mysql store with the given config.
func (o *DBOptions) NewClient() (*gorm.DB, error) {
	opts := gormx.Options{
		DriverName:            strings.ToLower(o.DBType),
		Host:                  o.Host,
		Port:                  o.Port,
		Username:              o.Username,
		Password:              o.Password,
		DBName:                o.Database,
		Config:                o.Config,
		MaxIdleConnections:    o.MaxIdleConnections,
		MaxOpenConnections:    o.MaxOpenConnections,
		MaxConnectionIdleTime: o.MaxConnectionIdleTime,
		MaxConnectionLifeTime: o.MaxConnectionLifeTime,
		LogLevel:              o.LogLevel,
		IsDebug:               o.IsDebug,
		TablePrefix:           o.TablePrefix,
	}

	return gormx.NewOnce(opts)
}
