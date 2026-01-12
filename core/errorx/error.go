package errorx

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agcodes"
	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agerrors"
)

type ErrorCodeInfo struct {
	code        string
	description string
	cause       string
	solution    string
}

func (e *ErrorCodeInfo) GetCode() string {
	return e.code
}

func (e *ErrorCodeInfo) Err() error {
	return e.newCoder(e.code, nil)
}

func (e *ErrorCodeInfo) Desc(args ...any) error {
	return e.newCoder(e.code, nil, args...)
}

func (e *ErrorCodeInfo) Detail(err any, args ...any) error {
	return e.newCoder(e.code, err, args...)
}

func (e *ErrorCodeInfo) newCoder(errCode string, err any, args ...any) error {
	desc := e.description
	if len(args) > 0 {
		desc = FormatDescription(desc, args...)
	}
	if err == nil {
		err = struct{}{}
	}

	coder := agcodes.New(errCode, desc, e.cause, e.solution, err, "")
	return agerrors.NewCode(coder)
}

// IsCode  判断是不是errorcode error
func IsCode(err error) bool {
	_, ok := err.(*agerrors.Error)
	return ok
}

// Is 比较两个err是不是同一个err
func Is(erra, errb error) bool {
	a, aok := erra.(*agerrors.Error)
	b, bok := errb.(*agerrors.Error)
	if !aok && !bok {
		return errors.Is(erra, errb)
	}
	if aok && bok {
		return a.Code() == b.Code()
	}
	return false
}

// FormatDescription replace the placeholder in coder.Description
// Example:
// Description: call service [service_name] api [api_name] error,
// args:  basic-service, create
// =>
// Description: call service [basic-service] api [create] error,
func FormatDescription(s string, args ...interface{}) string {
	if len(args) <= 0 {
		return s
	}
	re, _ := regexp.Compile("\\[\\w+\\]")
	result := re.ReplaceAll([]byte(s), []byte("[%v]"))
	return fmt.Sprintf(string(result), args...)
}

func FormatInvalidParamterDetailDescription(s string, args ...interface{}) string {
	if len(args) <= 0 {
		return s
	}
	re, _ := regexp.Compile("\\[\\w+\\]")
	result := re.ReplaceAll([]byte(s), []byte("%v"))
	return fmt.Sprintf(string(result), args...)
}

func FormatStringDescription(s string, args ...interface{}) string {
	if len(args) <= 0 {
		return s
	}
	re, _ := regexp.Compile("\\[\\w+\\]")
	result := re.ReplaceAll([]byte(s), []byte("%v"))
	return fmt.Sprintf(string(result), args...)
}
