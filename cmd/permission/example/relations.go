package example

import (
	"fmt"

	"github.com/kweaver-ai/idrm-go-frame/core/transport/rest/ginx"

	"github.com/gin-gonic/gin"
)

type MenuResource struct {
	ServiceName string `json:"service_name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
}

func ServiceACRegister(engine *gin.Engine, rs []*MenuResource) {
	serviceName := "data-view"
	address := fmt.Sprintf("/api/internal/%s/v1/ac", serviceName)
	engine.GET(address, func(c *gin.Context) {
		ginx.ResOKJson(c, rs)
	})
}

var menuResources []*MenuResource

func init() {

	menuResources = append(menuResources, &MenuResource{
		ServiceName: "data-view",
		Path:        "/api/data-view/v1/form_view",
		Method:      "GET",
		Resource:    "管理逻辑视图",
		Action:      "读取",
	})

	menuResources = append(menuResources, &MenuResource{
		ServiceName: "data-view",
		Path:        "/api/data-view/v1/form_view/:id",
		Method:      "POST",
		Resource:    "管理逻辑视图",
		Action:      "管理",
	})

	menuResources = append(menuResources, &MenuResource{
		ServiceName: "data-view",
		Path:        "/api/data-view/v1/form_view/:id",
		Method:      "POST",
		Resource:    "管理数据质量",
		Action:      "管理",
	})

}
