package gormx

import (
	"gorm.io/gorm/logger"
)

// Options defines optsions for mysql database.
type Options struct {
	DriverName            string `json:",default=mysql,options=mysql|postgres|sqlserver|clickhouse|oracle|dm8|mariadb|tidb|goldendb"`
	Host                  string `json:",default=localhost"` // address
	Port                  string `json:",default=3330"`      // port
	Config                string `json:",optional"`          // extra config such as mysql:charset=utf8mb4&parseTime=True&loc=Local或者postgres:sslmode=disable TimeZone=Asia/Shangh或者clickhouse:read_timeout=10&write_timeout=20 达梦开启compatibleMode=mysql
	DBName                string `json:",default=idrm_main"` // orm name
	Username              string `json:",default=root"`      // username
	Password              string `json:",default=root"`      // password
	MaxIdleConnections    int    `json:",default=5"`
	MaxOpenConnections    int    `json:",default=50"`
	MaxConnectionIdleTime int    `json:",default=10"`  // 秒
	MaxConnectionLifeTime int    `json:",default=500"` // unit 秒
	LogLevel              int
	Logger                logger.Interface
	IsDebug               bool
	TablePrefix           string
}
