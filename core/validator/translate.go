package validator

import (
	"fmt"
	"strings"

	"github.com/kweaver-ai/idrm-go-frame/core/enum"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// EnumTranslation add additional replacement, "{0}的值必须是{1}之一",
func EnumTranslation(tran ut.Translator, fe validator.FieldError) string {
	enumObject := fe.Param()
	all := enum.Values(enumObject)
	params := strings.Join(all, ",")
	t, err := tran.T(fe.Tag(), fe.Field(), params)
	if err != nil {
		fmt.Printf("警告: 翻译字段错误: %s", err)
		return fe.(error).Error()
	}
	return t
}
