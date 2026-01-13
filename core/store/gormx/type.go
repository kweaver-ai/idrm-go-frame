package gormx

import (
	"gitee.com/tdxmkf123/gorm-driver-dameng/dameng"
	//"github.com/ClickHouse/clickhouse-go"
	"gorm.io/driver/postgres"

	"gitee.com/tdxmkf123/gorm-driver-oracle/oracle"
	"gorm.io/driver/gaussdb"

	//"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

const (
	DriverMysql      string = "mysql"
	DriverPostgres   string = "postgres"
	DriverSqlserver  string = "sqlserver"
	DriverClickhouse string = "clickhouse"
	DriveDm          string = "dm8"
	DriveMariaBb     string = "mariadb"
	DriveTiBb        string = "tidb"
	DriveGoldenDB    string = "goldendb"
	DriveOracle      string = "oracle"
	DriverGaussDB    string = "gaussdb"
)

var opens = map[string]func(string) gorm.Dialector{
	DriverMysql:     mysql.Open,
	DriveMariaBb:    mysql.Open,
	DriveTiBb:       mysql.Open,
	DriveGoldenDB:   mysql.Open,
	DriverPostgres:  postgres.Open,
	DriverSqlserver: sqlserver.Open,
	//driverClickhouse: clickhouse.Open,
	DriveOracle:   oracle.Open,
	DriveDm:       dameng.Open,
	DriverGaussDB: gaussdb.Open,
}
