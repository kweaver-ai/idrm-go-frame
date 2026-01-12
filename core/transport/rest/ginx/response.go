package ginx

import (
	"net/http"

	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agcodes"
	"github.com/kweaver-ai/idrm-go-frame/core/errorx/agerrors"

	"github.com/gin-gonic/gin"
)

const StatusCode = "StatusCode"

type HttpError struct {
	Code        string      `json:"code"`
	Description string      `json:"description"`
	Solution    string      `json:"solution,omitempty"`
	Cause       string      `json:"cause,omitempty"`
	Detail      interface{} `json:"detail,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

// success Json Response
func ResOKJson(c *gin.Context, data interface{}) {

	if data == nil {
		data = gin.H{}
	}
	c.JSON(http.StatusOK, data)
}

// list Response
func ResList(c *gin.Context, list interface{}, totalCount int) {

	c.JSON(http.StatusOK, gin.H{
		"entries":     list,
		"total_count": totalCount,
	})

}
func ResBadRequestJson(c *gin.Context, err error) {
	ResErrJsonWithCode(c, http.StatusBadRequest, err)
}
func ResErrJsonWithCode(c *gin.Context, code int, err error) {
	c.Writer.WriteHeader(code)
	ResErrJson(c, err)
}

// failed Json Response
func ResErrJson(c *gin.Context, err error) {
	if value, exists := c.Get(StatusCode); exists {
		if code, ok := value.(int); ok && code >= 100 && code < 600 {
			c.Writer.WriteHeader(code)
		}
	}
	var code agcodes.Coder
	if err == nil {
		code = agcodes.CodeOK
	} else {
		code = agerrors.Code(err)
	}

	c.JSON(c.Writer.Status(), HttpError{
		Code:        code.GetErrorCode(),
		Description: code.GetDescription(),
		Solution:    code.GetSolution(),
		Cause:       code.GetCause(),
		Detail:      code.GetErrorDetails(),
	})
}

func AbortResponseWithCode(c *gin.Context, code int, err error) {
	c.Writer.WriteHeader(code)
	AbortResponse(c, err)
}
func AbortResponse(c *gin.Context, err error) {
	var code = agerrors.Code(err)
	if err == nil {
		code = agcodes.CodeNotAuthorized
	}
	c.AbortWithStatusJSON(c.Writer.Status(), HttpError{
		Code:        code.GetErrorCode(),
		Description: code.GetDescription(),
		Solution:    code.GetSolution(),
		Cause:       code.GetCause(),
		Detail:      code.GetErrorDetails(),
	})
}
