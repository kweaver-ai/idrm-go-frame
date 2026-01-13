package gormx

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func GetDriverName(db *gorm.DB) string {
	name := db.Dialector.Name()
	if strings.Contains(DriveDm, name) {
		return DriveDm
	}
	return name
}

// Field 实现mysql的Field功能
func Field(db *gorm.DB, args ...string) string {
	argsLen := len(args)
	if len(args) <= 2 {
		return ""
	}
	fieldName := args[0]
	order := args[1]

	placeHolderSlice := make([]string, len(args)-2)
	for i := 2; i < argsLen; i++ {
		placeHolderSlice[i-2] = "'%s'"
	}

	driver := GetDriverName(db)
	switch driver {
	case DriveMariaBb, DriverMysql:
		return fmt.Sprintf("Field(`%s`, %s) %s", fieldName, strings.Join(placeHolderSlice, ","), order)
	case DriveDm:
		return fmt.Sprintf(`POSITION("%s" in (%s)) %s`, fieldName, strings.Join(placeHolderSlice, ","), order)
	}
	return ""
}
