package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterALLToInternal(engine *gin.Engine) {
	for _, route := range engine.Routes() {
		path := insertString(route.Path, "/internal", 4)
		switch route.Method {
		case http.MethodGet:
			engine.GET(path, route.HandlerFunc)
		case http.MethodPost:
			engine.POST(path, route.HandlerFunc)
		case http.MethodPut:
			engine.PUT(path, route.HandlerFunc)
		case http.MethodPatch:
			engine.PATCH(path, route.HandlerFunc)
		case http.MethodDelete:
			engine.DELETE(path, route.HandlerFunc)
		}
	}
}
func insertString(original, insertion string, index int) string {
	if index < 0 || index > len(original) {
		return original
	}
	return original[:index] + insertion + original[index:]
}

// 内部接口自动注册，适用于Token透传中间件
func RegisterALLToInternalWithMiddleware(routes gin.RoutesInfo, internalRouter *gin.RouterGroup) {
	for _, route := range routes {
		path := insertString(route.Path, "/internal", 4)
		switch route.Method {
		case http.MethodGet:
			internalRouter.GET(path, route.HandlerFunc)
		case http.MethodPost:
			internalRouter.POST(path, route.HandlerFunc)
		case http.MethodPut:
			internalRouter.PUT(path, route.HandlerFunc)
		case http.MethodPatch:
			internalRouter.PATCH(path, route.HandlerFunc)
		case http.MethodDelete:
			internalRouter.DELETE(path, route.HandlerFunc)
		}
	}
}
