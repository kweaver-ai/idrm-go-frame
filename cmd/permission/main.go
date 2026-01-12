package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

var args = &CmdArg{}

type CmdArg struct {
	Path        string `json:"path"`
	ServiceName string `json:"service_name"`
	BaseRoute   string `json:"base_route"`
	Dest        string `json:"dest"`
	DestPackage string `json:"dest_package"`
}

func (c *CmdArg) Check() error {
	if c.Path == "" {
		return fmt.Errorf("'path' can't be empty")
	}
	if c.ServiceName == "" {
		return fmt.Errorf("'service_name' can't be empty")
	}
	if c.Dest == "" {
		return fmt.Errorf("'dest' can't be empty")
	}
	if !strings.HasSuffix(c.Dest, ".go") {
		return fmt.Errorf("'dest' must end with '.go'")
	}
	if c.DestPackage == "" {
		return fmt.Errorf("'dest_package' can't be empty")
	}
	return nil
}

func (c *CmdArg) NewProjectParser() *ProjectParser {
	return NewProjectParser(c.Path, c.ServiceName, c.BaseRoute, c.Dest, c.DestPackage)
}

func init() {
	flag.StringVar(&args.Path, "path", ".", "需要解析的项目或文件路径,必填， 例如: -path ./data-view")
	flag.StringVar(&args.ServiceName, "service-name", "", "服务名称，必填， 例如: data-view")
	flag.StringVar(&args.BaseRoute, "base-route", "", "基本路由地址，为空可不填， 例如：/api/data-view/v1")
	flag.StringVar(&args.Dest, "dest", "permission_resource.go", "权限资源关系代码路径,必填， 例如: -dest permission_resource.go")
	flag.StringVar(&args.DestPackage, "dest-package", "example", "权限资源关系代码的包名称， 例如: -dest example")
}

func main() {
	flag.Parse()
	if err := args.Check(); err != nil {
		fmt.Println(err.Error())
		return
	}
	parser := args.NewProjectParser()
	ans, err := parser.ReadAllReadAnnotation()
	if err != nil {
		fmt.Printf("读取注解失败:%v", err.Error())
		return
	}
	//生成代码
	if err = WriteTemplate(permissionResourceTemplate, ans, args.Dest); err != nil {
		fmt.Printf("根据注解生成代码失败:%v", err.Error())
	} else {
		absDest, _ := filepath.Abs(args.Dest)
		if absDest != "" {
			fmt.Printf("注解代码生成路径:%v", absDest)
		}
	}
	return
}
