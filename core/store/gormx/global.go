package gormx

import (
	"fmt"
	"gorm.io/gorm"
)

const GroupOptFlag = "GROUP_OPT_FLAG"

func GetDM8Variable(db *gorm.DB, k string) (value string, err error) {
	rawSQL := fmt.Sprintf("SELECT para_value FROM v$dm_ini WHERE para_name='%v';", k)
	err = db.Raw(rawSQL).Scan(&value).Error
	return value, err
}

func SetDM8Variable(db *gorm.DB, k, v string) error {
	rawSQL := fmt.Sprintf("SP_SET_PARA_VALUE(1,'%v',%v);", k, v)
	return db.Raw(rawSQL).Error
}

func StoreDM8Variable(db *gorm.DB, k, v string) func() {
	//取到原值
	sourceValue, err := GetDM8Variable(db, k)
	if err != nil {
		panic(err)
	}
	//设置值
	if err = SetDM8Variable(db, k, v); err != nil {
		panic(err)
	}
	//使用defer还原
	return func() {
		if err := SetDM8Variable(db, k, sourceValue); err != nil {
			panic(err)
		}
	}
}

func Restore(f func()) {
	f()
}

func SetDM8Compatible(db *gorm.DB) func() {
	return StoreDM8Variable(db, GroupOptFlag, "1")
}
