package rest

import (
	"net/http"

	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agcodes"
	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agerrors"

	"github.com/gin-gonic/gin"
)

type HttpError struct {
	Code        string      `json:"code"  example:"task-center.Public.InternalError"`     //返回错误码，格式: 服务名.模块.错误
	Description string      `json:"description"  example:"内部错误" `                         //错误描述
	Solution    string      `json:"solution" extensions:"!x-omitempty" example:"请联系内部人员"` //错误处理办法
	Cause       string      `json:"cause"  extensions:"!x-omitempty"  example:"错误原因"`     //错误原因
	Detail      interface{} `json:"detail,omitempty"  extensions:"!x-omitempty"`          //错误详情, 一般是json对象
}

// success Json Response
func ResOKJson(c *gin.Context, data interface{}) {

	c.JSON(http.StatusOK, data)
}

// failed Json Response
func ResErrJson(c *gin.Context, err error) {
	var (
		code       = agerrors.Code(err)
		statusCode = 400
	)
	if err != nil {
		if code == agcodes.CodeNil {
			code = agcodes.CodeInternalError
		}
	} else if c.Writer.Status() > 0 && c.Writer.Status() != http.StatusOK {
		//switch c.Writer.Status() {
		//case http.StatusNotFound:
		//    code = agcodes.CodeNotFound
		//case http.StatusForbidden:
		//    code = agcodes.CodeNotAuthorized
		//
		//default:
		//    code = agcodes.CodeInternalError
		//}
		statusCode = c.Writer.Status()
	} else {
		code = agcodes.CodeOK
		statusCode = 200
	}

	c.JSON(statusCode, HttpError{
		Code:        code.GetErrorCode(),
		Description: code.GetDescription(),
		Solution:    code.GetSolution(),
		Cause:       code.GetCause(),
		Detail:      code.GetErrorDetails(),
	})
}
