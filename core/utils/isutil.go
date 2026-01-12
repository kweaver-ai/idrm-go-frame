package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/kweaver-ai/idrm-go-frame/core/common"
)

func IsBlank(value string) bool {
	v := strings.TrimSpace(value)
	if v != "" {
		return false
	}
	return true
}

func IsNotBlank(value string) bool {
	v := strings.TrimSpace(value)
	if v != "" {
		return true
	}
	return false
}

// IsNumber - 是否为数字
func IsNumber(value any) bool {
	if value == nil {
		return false
	}
	return regexp.MustCompile(`^[0-9]+$`).MatchString(ToString(value))
}

// IsFloat - 是否为浮点数
func IsFloat(value any) bool {
	if value == nil {
		return false
	}
	return regexp.MustCompile(`^[0-9]+(.[0-9]+)?$`).MatchString(ToString(value))
}

// IsArrayContain -数组是否包含该值
func IsArrayContain[T comparable](value T, array []T) bool {
	for _, item := range array {
		if item == value {
			return true
		}
	}
	return false
}

// IsInUpChar - 是否包含大写字母
func IsInUpChar(word string) bool {
	head := word[:1]
	return isInChar(head, 'A', 'Z')
}
func IsInLoweChar(word string) bool {
	tail := word[1:]
	return isInChar(tail, 'a', 'z')
}

func isInChar(s string, first, last byte) bool {
	for i := range s {
		if !(first <= s[i] && s[i] <= last) {
			return false
		}
	}
	return true
}

func IsBlankArr(values []string) bool {
	if len(values) > 0 {
		return false
	}
	return true
}

func IsNotBlankArr(values []string) bool {
	if len(values) > 0 {
		return true
	}
	return false
}

func IsConfSuffix(suffix string) bool {
	if suffix != common.JsonType && suffix != common.TomlType && suffix != common.YamlType && suffix != common.YmlType {
		return false
	}
	return true
}

func IsBlankMap(maps map[string]interface{}) bool {
	if len(maps) > 0 {
		return false
	}
	return true
}

func IsBlankTime(t *time.Time) bool {
	if t == nil {
		return true
	}
	return false
}

func ToString(i interface{}) (str string) {
	switch i.(type) {
	case string:
		str = i.(string)
	case int:
		str = strconv.Itoa(i.(int))
	case int8:
		str = fmt.Sprint(i.(int8))
	case int16:
		str = fmt.Sprint(i.(int16))
	case int32:
		str = string(i.(int32))
	case int64:
		str = strconv.FormatInt(i.(int64), 10)
	case uint:
		str = strconv.Itoa(i.(int))
	case uint8:
		str = string(i.(uint8))
	case uint16:
		str = fmt.Sprint(i.(uint16))
	case uint32:
		str = fmt.Sprint(i.(uint32))
	case uint64:
		str = strconv.FormatUint(i.(uint64), 10)
	case float32:
		str = fmt.Sprintf("%f", i.(float32))
	case float64:
		str = strconv.FormatFloat(i.(float64), 'f', -1, 32)
	case time.Time:
		str = i.(time.Time).Format("2006-01-02 15:04:05")
	case []byte:
		b := i.([]byte)
		str = *(*string)(unsafe.Pointer(&b))
	case error:
		str = i.(error).Error()
	default:
		panic(common.ErrorTypeSupported)
	}
	return
}
