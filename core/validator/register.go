package validator

import (
	"github.com/kweaver-ai/idrm-go-frame/core/validator/service"

	"github.com/gin-gonic/gin/binding"
)

var customValidatorObjects []service.CustomValidatorObject

func init() {
	registerCustomValidatorObjects()
}

func SetupValidator() error {
	customV := service.NewCustomValidator().(*service.CustomValidator)
	binding.Validator = customV

	if err := service.InitTrans(customV.Validate, customValidatorObjects); err != nil {
		panic(err.Error())
	}
	return nil
}

func RegisterCustomValidatorObject(objs ...service.CustomValidatorObject) {
	customValidatorObjects = append(customValidatorObjects, objs...)
}

func registerCustomValidatorObjects() {
	customValidatorObjects = []service.CustomValidatorObject{
		{
			Tag:           "verifyEnum",
			ValidatorFunc: VerifyEnum,
			Trans: map[string]string{
				"zh": "{0}的值必须是[{1}]其中之一",
				"en": "{0}的值必须是[{1}]其中之一",
			},
			TranslationFunc: EnumTranslation,
		},
		{
			Tag:           "ValidateSnowflakeID",
			ValidatorFunc: ValidateSnowflakeID,
			Trans: map[string]string{
				"zh": "{0}只支持雪花Id的正整数输入",
				"en": "{0}只支持雪花Id的正整数输入",
			},
		},
		{
			Tag:           "VerifyXssString",
			ValidatorFunc: VerifyXssString,
			Trans: map[string]string{
				"zh": "{0}不支持insert、drop、delete等输入",
				"en": "{0}不支持insert、drop、delete等输入",
			},
		},
		{
			Tag:           "VerifyHost",
			ValidatorFunc: VerifyHost,
			Trans: map[string]string{
				"zh": "{0}不符合规范",
				"en": "{0}不符合规范",
			},
		},
		{
			Tag:           "TrimSpace",
			ValidatorFunc: TrimSpace,
			Trans:         map[string]string{},
		},
	}
}
