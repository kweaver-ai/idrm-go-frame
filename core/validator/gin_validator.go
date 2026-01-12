package validator

import (
	"reflect"
	"strings"

	"github.com/kweaver-ai/idrm-go-frame/core/validator/service"

	"github.com/gin-gonic/gin"
)

const (
	ParamTypeStructTag = "param_type"

	ParamTypeUri   = "path"
	ParamTypeQuery = "query"
	ParamTypeBody  = "body"

	ParamTypeBodyContentTypeJson = "json"
	ParamTypeBodyContentTypeForm = "form"
)

func Valid[T any](c *gin.Context) (*T, error) {
	t := new(T)
	value := reflect.ValueOf(t)

	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value = reflect.New(value.Elem().Type())
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		panic("req param T must struct")
	}

	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := value.Field(i)

		if !fieldType.Anonymous {
			continue
		}

		if fieldValue.Kind() != reflect.Struct {
			panic("struct field must struct")
		}

		paramType := fieldType.Tag.Get(ParamTypeStructTag)
		if len(paramType) < 1 {
			continue
		}

		idx := strings.Index(paramType, "=")
		var p string
		if idx > 0 {
			p = paramType[idx+1:]
			paramType = paramType[:idx]
		}

		var validatorFunc func(c *gin.Context, v interface{}) (bool, error)
		switch paramType {
		case ParamTypeUri:
			validatorFunc = service.BindUriAndValid

		case ParamTypeQuery:
			validatorFunc = service.BindQueryAndValid

		case ParamTypeBody:
			if len(p) < 1 {
				p = ParamTypeBodyContentTypeJson
			}
			switch p {
			case ParamTypeBodyContentTypeJson:
				validatorFunc = service.BindJsonAndValid

			case ParamTypeBodyContentTypeForm:
				validatorFunc = service.BindFormAndValid

			default:
				panic("not support param type")
			}
		default:
			panic("not support param type")
		}

		if _, err := validatorFunc(c, fieldValue.Addr().Interface()); err != nil {
			return nil, err
		}
	}
	return value.Addr().Interface().(*T), nil
}

// BindAndValid bind data from form and  validate
func BindAndValid(c *gin.Context, v interface{}) (bool, error) {
	return service.BindAndValid(c, v)
}

// BindFormAndValid parse and validate parameters in form-data
func BindFormAndValid(c *gin.Context, v interface{}) (bool, error) {
	return service.BindFormAndValid(c, v)
}

// BindQueryAndValid parse and validate parameters in query
func BindQueryAndValid(c *gin.Context, v interface{}) (bool, error) {
	return service.BindQueryAndValid(c, v)
}

// BindUriAndValid parse and validate parameters in uri
func BindUriAndValid(c *gin.Context, v interface{}) (bool, error) {
	return service.BindUriAndValid(c, v)
}

func BindJsonAndValid(c *gin.Context, v interface{}) (bool, error) {
	return service.BindJsonAndValid(c, v)
}
