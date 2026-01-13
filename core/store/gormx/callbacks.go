package gormx

import (
	"strings"

	"gorm.io/gorm"
)

var callbackMap map[string]func(*gorm.DB)

func init() {
	callbackMap = make(map[string]func(*gorm.DB))
	callbackMap[DriveDm] = registerDMCallback
	callbackMap[DriveOracle] = registerOracleCallback
}

func RegisterCallback(client *gorm.DB) {
	driverName := client.Name()
	if driverName == "dm" {
		driverName = "dm8"
	}
	register, ok := callbackMap[driverName]
	if ok {
		register(client)
	}
}

func registerDMCallback(client *gorm.DB) {
	client.Callback().Query().Before("gorm:query").Register("custom:query_fix", fixMariaDBSQL)
	client.Callback().Raw().Before("gorm:raw").Register("custom:raw_fix", fixMariaDBSQL)
	client.Callback().Row().Before("gorm:row").Register("custom:row_fix", fixMariaDBSQL)
}

func registerOracleCallback(client *gorm.DB) {
	client.Callback().Query().Before("gorm:query").Register("custom:query_fix", fixMariaDBSQL)
	client.Callback().Raw().Before("gorm:raw").Register("custom:raw_fix", fixMariaDBSQL)
	client.Callback().Row().Before("gorm:row").Register("custom:row_fix", fixMariaDBSQL)
}

func fixMariaDBSQL(db *gorm.DB) {
	rawSQL := db.Statement.SQL.String()
	if rawSQL == "" {
		return
	}
	db.Statement.SQL.Reset()
	db.Statement.SQL.WriteString(ReplaceInvalid(rawSQL))
}

func ReplaceInvalid(s string) string {
	s = strings.ReplaceAll(s, "`", "\"")
	s = strings.ReplaceAll(s, "STR_TO_DATE", "TO_DATE")
	return s
}

func RawCount(db *gorm.DB) (total int64, err error) {
	rawSQL := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Count(&total)
	})
	err = db.Raw(rawSQL).Scan(&total).Error
	return total, err
}

func RawScan[T any](db *gorm.DB) (ts []T, err error) {
	rawSQL := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&ts)
	})
	err = db.Raw(rawSQL).Scan(&ts).Error
	return ts, err
}
func RawScanObj[T any](db *gorm.DB) (ts T, err error) {
	rawSQL := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Scan(&ts)
	})
	err = db.Raw(rawSQL).Scan(&ts).Error
	return ts, err
}
func RawFirst[T any](db *gorm.DB) (data T, err error) {
	rawSQL := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.First(&data)
	})
	err = db.Raw(rawSQL).Scan(&data).Error
	return data, err
}
