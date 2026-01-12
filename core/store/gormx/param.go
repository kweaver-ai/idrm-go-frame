package gormx

import (
	"net/url"
	"strings"
)

func BuildQueryParams(queryParams url.Values) (query url.Values, pageModel PageModel) {
	params := url.Values{}
	var arr []string
	var page PageModel
	for key, values := range queryParams {
		if IsBlankArr(values) {
			continue
		}
		switch key {
		case "sort":
			params["sort"] = values
		case "select":
			params["select"] = values
		case "omit":
			params["omit"] = values
		case "gcond":
			params["gcond"] = values
		case "pageNo":
			pageNo, _ := StrToInt(values[0])
			page.PageNo = pageNo
		case "pageSize":
			pageSize, _ := StrToInt(values[0])
			page.PageSize = pageSize
		default:
			for _, value := range values {
				columKey := parseColumn(DealNoStringType(key))
				arr = append(arr, columKey+"="+DealNoStringType(SqlFilter(value, false)))
			}
		}
	}
	if len(arr) > 0 {
		params["q"] = arr
	}
	return params, page
}

func parseColumn(key string) string {
	key = strings.TrimSpace(key)
	if len(key) == 0 {
		return ""
	}
	var str strings.Builder
	strArr := strings.Split(key, "")
	for index, value := range strArr {
		if IsInUpChar(value) {
			if index == 0 {
				value = strings.ToLower(value)
			} else {
				value = "_" + strings.ToLower(value)
			}
		}
		str.WriteString(value)
	}
	return str.String()
}
