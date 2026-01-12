package  {{.PackageName}}

import (
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

func ServiceACRegister(engine *gin.Engine, rs []*MenuResource)  {
    serviceName := "{{- .ServiceName}}"
   	address := fmt.Sprintf("/api/internal/%s/v1/ac", serviceName)
   	engine.GET(address, func(c *gin.Context) {
   		ginx.ResOKJson(c, rs)
   	})
}

var menuResources []*MenuResource

func init(){

    {{ range .Annos }}
    menuResources = append(menuResources, &MenuResource{
    	ServiceName:    "{{- .ServiceName}}",
    	Path:           "{{- .Path}}",
    	Method:         "{{- .Method}}",
    	Resource:       "{{- .Resource}}",
    	Action:         "{{- .Action}}",
    })
    {{ end}}
}
