package service

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	universal_translator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"strings"
)

// BindAndValid bind data from form and  validate
func BindAndValid(c *gin.Context, v interface{}) (bool, error) {
	var err error
	b := binding.Default(c.Request.Method, c.ContentType())
	switch b {
	case binding.Query:
		b = customQuery

	case binding.Form:
		b = customForm

	case binding.FormMultipart:
		b = customFormMultipart
	}

	err = b.Bind(c.Request, v)
	if err != nil {
		validatorErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return false, err
		}

		return false, genStructError(validatorErrors.Translate(getTrans(c)))
	}

	return true, nil
}

// BindFormAndValid parse and validate parameters in form-data
func BindFormAndValid(c *gin.Context, v interface{}) (bool, error) {
	if err := c.ShouldBindWith(v, customForm); err != nil {
		validatorErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return false, err
		}
		return false, genStructError(validatorErrors.Translate(getTrans(c)))
	}
	return true, nil
}

// BindQueryAndValid parse and validate parameters in query
func BindQueryAndValid(c *gin.Context, v interface{}) (bool, error) {
	if err := c.ShouldBindWith(v, customQuery); err != nil {
		validatorErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return false, err
		}
		return false, genStructError(validatorErrors.Translate(getTrans(c)))
	}
	return true, nil
}

// BindUriAndValid parse and validate parameters in uri
func BindUriAndValid(c *gin.Context, v interface{}) (bool, error) {
	err := c.ShouldBindUri(v)
	if err != nil {
		validatorErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return false, err
		}

		return false, genStructError(validatorErrors.Translate(getTrans(c)))
	}

	return true, nil
}

func BindJsonAndValid(c *gin.Context, v interface{}) (bool, error) {
	err := c.ShouldBindJSON(v)
	if err != nil {
		if validatorErrors, ok := err.(validator.ValidationErrors); ok {
			return false, genStructError(validatorErrors.Translate(getTrans(c)))
		}
		if isBindError, err1 := IsBindError(c, err); isBindError {
			return false, err1
		}
		if jsonUnmarshalTypeError, ok := err.(*json.UnmarshalTypeError); ok {
			var validErrors ValidErrors
			validErrors = append(validErrors, &ValidError{
				Key:     jsonUnmarshalTypeError.Field,
				Message: "请输入符合要求的数据类型和数据范围",
			})
			return false, validErrors
		}

		if jsonUnsupportedTypeError, ok := err.(*json.UnsupportedTypeError); ok {
			var validErrors ValidErrors
			validErrors = append(validErrors, &ValidError{
				Key:     jsonUnsupportedTypeError.Type.Name(),
				Message: "不支持的json数据类型",
			})
			return false, validErrors
		}

		if jsonUnsupportedValueError, ok := err.(*json.UnsupportedValueError); ok {
			var validErrors ValidErrors
			validErrors = append(validErrors, &ValidError{
				Key:     jsonUnsupportedValueError.Str,
				Message: "不支持的json数据值",
			})
			return false, validErrors
		}

		return false, err
	}
	return true, nil
}

// BindStructAndValid parse and validate parameters in uri
func BindStructAndValid(v interface{}) (bool, error) {
	err := binding.Validator.ValidateStruct(v)
	if err != nil {
		validatorErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return false, err
		}
		trans, _ := uniTrans.FindTranslator("zh")
		return false, genStructError(validatorErrors.Translate(trans))
	}

	return true, nil
}

// genStructError remove struct name in validate error, then return ValidErrors
func genStructError(fields map[string]string) ValidErrors {
	var errs ValidErrors
	// removeTopStruct 去除字段名中的结构体名称标识
	// refer from:https://github.com/go-playground/validator/issues/633#issuecomment-654382345
	for field, err := range fields {
		errs = append(errs, &ValidError{
			//Key:     field[strings.LastIndex(field, ".")+1:],
			Key:     field[strings.Index(field, ".")+1:],
			Message: err,
		})
	}
	return errs
}

func getLocale(c *gin.Context) []string {
	acceptLanguage := c.GetHeader("Accept-Language")
	ret := make([]string, 0)
	for _, lang := range strings.Split(acceptLanguage, ",") {
		if len(lang) == 0 {
			continue
		}

		ret = append(ret, strings.SplitN(lang, ";", 2)[0])
	}

	return ret
}

func getTrans(c *gin.Context) universal_translator.Translator {
	locales := getLocale(c)

	trans, _ := uniTrans.FindTranslator(locales...)
	return trans
}
