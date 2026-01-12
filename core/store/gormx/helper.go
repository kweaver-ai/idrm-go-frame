package gormx

import (
	"html"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var reStr = `(?:')|(?:\%)|(?:\\)|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`

type PageModel struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

func SqlFilter(matchStr string, exactly bool) string {
	re, err := regexp.Compile(reStr)
	if err != nil {
		log.Println(err)
		return ``
	}
	return re.ReplaceAllStringFunc(matchStr, func(dest string) string {
		if dest == `\` && !exactly {
			dest = strings.ReplaceAll(dest, dest, `\\\`+dest)
		} else {
			dest = strings.ReplaceAll(dest, dest, `\`+dest)
		}
		return dest
	})
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

func StrToInt(i interface{}) (num int, err error) {
	switch i.(type) {
	case string:
		num, err = strconv.Atoi(i.(string))
	}
	return
}

func DealNoStringType(value string) string {
	valueStr := html.EscapeString(value)
	return valueStr
}
