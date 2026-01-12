package gormx

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/utils"

	"github.com/acmestack/gorm-plus/gplus"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	ENV_RDSUSER   = "DB_USERNAME"
	ENV_RDSPASS   = "DB_PASSWORD"
	ENV_RDSHOST   = "DB_HOST"
	ENV_RDSPORT   = "DB_PORT"
	ENV_RDSDBNAME = "DB_NAME"
)

// New create a new gorm db instance with the given options.
func New(e Options) (*gorm.DB, error) {
	dbType := os.Getenv("DB_TYPE")
	if utils.IsNotBlank(dbType) {
		dbType = strings.ToLower(dbType)
		e.DriverName = dbType
	}
	driver, ok := opens[e.DriverName]
	if !ok {
		return nil, errors.New("orm dialect is not supported")
	}
	dsn := dbHandler(e)
	gdb, err := gorm.Open(driver(dsn), &gorm.Config{
		PrepareStmt: true,
		QueryFields: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,          //表名后面不加s
			TablePrefix:   e.TablePrefix, // 表前缀
		},
		Logger: e.Logger,
	})
	if err != nil {
		return nil, err
	}
	if e.IsDebug {
		gdb = gdb.Debug()
	}
	//设置回调
	RegisterCallback(gdb)
	//设置链接配置
	sqlDb, sqlErr := gdb.DB()
	if sqlErr != nil {
		return nil, sqlErr
	}
	sqlDb.SetMaxIdleConns(e.MaxIdleConnections)                                    //设置最大的空闲连接数
	sqlDb.SetMaxOpenConns(e.MaxOpenConnections)                                    //设置最大连接数
	sqlDb.SetConnMaxLifetime(time.Duration(e.MaxConnectionLifeTime) * time.Second) //可重用链接得最大时间
	sqlDb.SetConnMaxIdleTime(time.Duration(e.MaxConnectionIdleTime) * time.Second) //越短连接过期的次数就会越频繁
	sqlDb.Ping()
	gplus.Init(gdb)
	return gdb, nil
}

func postgresDSN(e Options) string {
	var config = "sslmode=disable TimeZone=Asia/Shanghai"
	if utils.IsNotBlank(e.Config) {
		config = e.Config
	}
	dSn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s %s", e.Host, e.Username, e.Password, e.DBName, e.Port, config)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dSn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s %s",
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			os.Getenv(ENV_RDSDBNAME),
			os.Getenv(ENV_RDSHOST),
			config)
	}
	return dSn
}

func mysqlDSN(e Options) string {
	var config = "charset=utf8mb4&parseTime=True&loc=Local"
	if utils.IsNotBlank(e.Config) {
		config = e.Config
	}
	dSn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", e.Username, e.Password, e.Host, e.Port, e.DBName, config)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dSn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			os.Getenv(ENV_RDSHOST),
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSDBNAME),
			config)
	}
	return dSn
}

func sqlServerDSN(e Options) string {
	dSn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?orm=%s", e.Username, e.Password, e.Host, e.Port, e.DBName)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dSn = fmt.Sprintf("sqlserver://%s:%s@%s:%s?orm=%s",
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			os.Getenv(ENV_RDSHOST),
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSDBNAME))
	}
	return dSn
}

func clickHouseDSN(e Options) string {
	var config = "read_timeout=10&write_timeout=20"
	if utils.IsNotBlank(e.Config) {
		config = e.Config
	}
	dSn := fmt.Sprintf("tcp://%s:%s?orm=%s&username=%s&password=%s&%s", e.Host, e.Port, e.DBName, e.Username, e.Password, config)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dSn = fmt.Sprintf("tcp://%s:%s?orm=%s&username=%s&password=%s&%s",
			os.Getenv(ENV_RDSHOST),
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSDBNAME),
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			config)
	}
	return dSn
}

func dmDSN(e Options) string {
	var config = "timeout=20s&autocommit=true&readTimeout=10s&genKeyNameCase=2&doSwitch=1"
	if utils.IsNotBlank(e.Config) {
		config = e.Config
	}
	// dm://sysdba:dameng123!@193.100.100.221:5236?autoCommit=true
	dSn := fmt.Sprintf("dm://%s:%s@%s:%s?schema=%s&%s", e.Username, e.Password, e.Host, e.Port, e.DBName, config)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dSn = fmt.Sprintf("dm://%s:%s@%s:%s?schema=%s&%s",
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			os.Getenv(ENV_RDSHOST),
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSDBNAME),
			config)
	}
	return dSn
}

func oracleDSN(e Options) string {
	// ZTK/sirc1234@193.100.100.43:1521/ORCL
	dSn := fmt.Sprintf("%s/%s@%s:%s/%s", e.Username, e.Password, e.Host, e.Port, e.DBName)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dSn = fmt.Sprintf("%s/%s@%s:%s/%s",
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			os.Getenv(ENV_RDSHOST),
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSDBNAME))
	}
	return dSn
}

func tidbDSN(e Options) string {
	var config = "charset=utf8mb4&parseTime=True&loc=Local"
	if utils.IsNotBlank(e.Config) {
		config = e.Config
	}
	dSn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", e.Username, e.Password, e.Host, e.Port, e.DBName, config)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dSn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			os.Getenv(ENV_RDSHOST),
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSDBNAME),
			config)
	}
	return dSn
}

func gaussDBDSN(e Options) string {
	config := "sslmode=disable TimeZone=Local"
	if utils.IsNotBlank(e.Config) {
		config = e.Config
	}
	dsnStr := "host=%s user=%s password=%s dbname=%s port=%s %s"
	dsn := fmt.Sprintf(dsnStr, e.Host, e.Username, e.Password, e.DBName, e.Port, config)
	user, _ := os.LookupEnv(ENV_RDSUSER)
	if utils.IsNotBlank(user) {
		dsn = fmt.Sprintf(dsnStr,
			os.Getenv(ENV_RDSUSER),
			os.Getenv(ENV_RDSPASS),
			os.Getenv(ENV_RDSHOST),
			os.Getenv(ENV_RDSPORT),
			os.Getenv(ENV_RDSDBNAME),
			config,
		)
	}
	return dsn
}

func dbHandler(e Options) string {
	var dsn string
	switch e.DriverName {
	case DriverMysql:
		dsn = mysqlDSN(e)
	case DriverPostgres:
		dsn = postgresDSN(e)
	case DriverSqlserver:
		dsn = sqlServerDSN(e)
	case DriverClickhouse:
		dsn = clickHouseDSN(e)
	case DriveDm:
		dsn = dmDSN(e)
	case DriveTiBb:
		dsn = tidbDSN(e)
	case DriveOracle:
		dsn = oracleDSN(e)
	case DriverGaussDB:
		dsn = gaussDBDSN(e)
	default:
		dsn = mysqlDSN(e)
	}
	return dsn
}

func getLevel(logMode string) logger.LogLevel {
	var level logger.LogLevel
	switch logMode {
	case "info":
		level = logger.Info
	case "warn":
		level = logger.Warn
	case "error":
		level = logger.Error
	default:
		level = logger.Error
	}
	return level
}

// 自定义一个 Writer
type writer struct {
	logger.Writer
}

func (l writer) Printf(message string, data ...any) {
	log.Println(message, data)
	// TODO上报日志到AR
}
