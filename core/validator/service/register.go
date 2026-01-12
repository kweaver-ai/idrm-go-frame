package service

import (
	"errors"
	"reflect"
	"strings"

	"github.com/kweaver-ai/idrm-go-frame/core/telemetry/log"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uniTrans *ut.UniversalTranslator
)

type CustomValidatorObject struct {
	Tag                      string
	ValidatorFunc            validator.Func
	CallValidationEvenIfNull bool
	Trans                    map[string]string
	TranslationFunc          validator.TranslationFunc
}

func registerCustomerValidationAndTranslation(v *validator.Validate, customerValidators []CustomValidatorObject) error {
	for _, customerValidator := range customerValidators {
		if len(customerValidator.Tag) == 0 {
			err := errors.New("tag is empty")
			log.Errorf("failed to customer validator, err: %v", err)
			return err
		}
		if customerValidator.ValidatorFunc == nil && len(customerValidator.Trans) == 0 {
			err := errors.New("customer validator func is nil")
			log.Errorf("failed to customer validator, err: %v", err)
			return err
		}

		if customerValidator.ValidatorFunc != nil {
			err := v.RegisterValidation(customerValidator.Tag, customerValidator.ValidatorFunc, customerValidator.CallValidationEvenIfNull)
			if err != nil {
				log.Errorf("failed to register customer validation, tag: %v, err: %v", customerValidator.Tag, err)
				return err
			}
		}

		for loc, msg := range customerValidator.Trans {
			tran, found := uniTrans.GetTranslator(loc)
			if !found {
				log.Warnf("no register locale translator, locale: %v", loc)
				continue
			}

			tranFunc := customerValidator.TranslationFunc
			if tranFunc == nil {
				tranFunc = translate
			}

			err := v.RegisterTranslation(customerValidator.Tag, tran, registerTranslator(customerValidator.Tag, msg), tranFunc)
			if err != nil {
				log.Errorf("failed to register customer translation, tag: %v, locale: %v, err: %v", customerValidator.Tag, loc, err)
				return err
			}
		}
	}

	return nil
}

func InitTrans(v *validator.Validate, customerValidators []CustomValidatorObject) error {
	zhT := zh.New()
	uniTrans = ut.New(zhT, zhT, en.New())
	enTran, _ := uniTrans.GetTranslator("en")
	zhTran, _ := uniTrans.GetTranslator("zh")

	err := enTranslations.RegisterDefaultTranslations(v, enTran)
	if err != nil {
		log.Errorf("failed to register en translations, err: %v", err)
		return err
	}

	err = zhTranslations.RegisterDefaultTranslations(v, zhTran)
	if err != nil {
		log.Errorf("failed to register zh translations, err: %v", err)
		return err
	}

	v.RegisterTagNameFunc(registerTagName)

	return registerCustomerValidationAndTranslation(v, customerValidators)
}

// registerTranslator 为自定义字段添加翻译功能
func registerTranslator(tag string, msg string, overrides ...bool) validator.RegisterTranslationsFunc {
	return func(trans ut.Translator) error {
		override := false
		if len(overrides) > 0 {
			override = overrides[0]
		}

		if err := trans.Add(tag, msg, override); err != nil {
			return err
		}
		return nil
	}
}

// translate 自定义字段的翻译方法
func translate(trans ut.Translator, fe validator.FieldError) string {
	msg, err := trans.T(fe.Tag(), fe.Field())
	if err != nil {
		log.Warnf("warning: error translating FieldError: %s", err)
		return fe.Error()
	}

	return msg
}

func registerTagName(field reflect.StructField) string {
	var name string
	for _, tagName := range []string{"name", "uri", "form", "json"} {
		name = FindTagName(field, tagName)
		if len(name) > 0 {
			return name
		}
	}

	return strings.ToLower(field.Name)
}

// FindTagName find tag value in structField c
func FindTagName(c reflect.StructField, tagName string) string {
	tagValue := c.Tag.Get(tagName)
	ts := strings.Split(tagValue, ",")
	if tagValue == "" || tagValue == "omitempty" || ts[0] == "" {
		return c.Name
	}
	if tagValue == "-" {
		return ""
	}
	if len(ts) == 1 {
		return tagValue
	}
	return ts[0]
}

func GetUniTrans() *ut.UniversalTranslator {
	return uniTrans
}
