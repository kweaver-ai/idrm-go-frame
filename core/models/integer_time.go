package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/utils"
)

type IntegerTime struct {
	time.Time
}

func NowIntegerTime() *IntegerTime {
	return &IntegerTime{Time: time.Now()}
}

func (t IntegerTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("0"), nil
	}

	// return StringToBytes(fmt.Sprintf("\"%s\"", t.Format(constant.LOCAL_TIME_FORMAT))), nil
	return utils.StringToBytes(fmt.Sprintf("%d", t.UnixMilli())), nil
}

func (t *IntegerTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		return nil
	}

	// str = strings.Trim(str, "\"")
	// val, err := time.Parse(constant.LOCAL_TIME_FORMAT, str)
	// *t = Time{val}
	ts, err := strconv.ParseInt(str, 10, 64)
	*t = IntegerTime{time.UnixMilli(ts)}
	return err
}

func (t *IntegerTime) Scan(value interface{}) error {
	val, ok := value.(time.Time)
	if ok {
		*t = IntegerTime{val}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", value)
}

func (t IntegerTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}
