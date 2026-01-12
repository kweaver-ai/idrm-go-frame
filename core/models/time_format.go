package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/utils"
)

const (
	LOCAL_TIME_FORMAT = "2006-01-02 15:04:05"
)

type StringTime struct {
	time.Time
}

func NowStringTime() *StringTime {
	return &StringTime{Time: time.Now()}
}

func (t StringTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte{}, nil
	}

	return utils.StringToBytes(fmt.Sprintf("\"%s\"", t.Format(LOCAL_TIME_FORMAT))), nil
}

func (t *StringTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		return nil
	}

	str = strings.Trim(str, "\"")
	val, err := time.Parse(LOCAL_TIME_FORMAT, str)
	*t = StringTime{val}
	return err
}

func (t *StringTime) Scan(value interface{}) error {
	val, ok := value.(time.Time)
	if ok {
		*t = StringTime{val}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", value)
}

func (t StringTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}
