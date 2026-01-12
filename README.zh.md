# idrm-go-frame

[English](README.md)

面向服务端应用的 Go 基础框架，提供应用生命周期、配置、日志、传输、注册发现、存储与基础工具能力。

## 项目简介
`idrm-go-frame` 是一个 Go 服务框架基座，聚焦在“可复用的基础能力”，包含服务启动与生命周期管理、配置管理、日志、HTTP/Kafka 传输、注册发现、存储与工具库等模块。

## 特性
- 应用生命周期管理（启动、优雅停机、服务注册）。
- 配置模块（多源加载、格式解析、热更新）。
- 日志模块（基于 zap，多输出、多级别配置）。
- 传输层：REST（Gin）与 Kafka（kafkax / kq）。
- 观测能力：telemetry 日志与 trace 封装。
- 存储与缓存：gormx、redis 工具、分布式锁等。
- 工具库：enum、encoding、syncx、utils 等。

## 目录结构
- `app.go`：应用生命周期与运行入口。
- `core/config`：配置加载与解析。
- `core/logx`：日志框架与配置。
- `core/transport/rest`：REST 服务封装（Gin）。
- `core/transport/mq/kafkax`：Kafka 消费/生产封装。
- `core/telemetry`：日志与 trace 组件。
- `core/store` / `core/redis_tool`：存储与缓存工具。
- `docs/`：组件与使用示例文档。

## 快速开始
```go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	idrm "github.com/kweaver-ai/idrm-go-frame"
	"github.com/kweaver-ai/idrm-go-frame/core/transport/rest"
)

func main() {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	srv := rest.NewServer(r, rest.Address(":8080"))
	app := idrm.New(idrm.Name("demo"), idrm.Server(srv))

	if err := app.Run(); err != nil {
		log.Fatalf("app run error: %v", err)
	}
}
```

## 文档
- 配置组件：`docs/component/config.md`
- 日志组件：`docs/component/logger.md`
- 枚举组件：`docs/component/enum.md`
- Kafka 使用：`docs/transport/kafkax.md`

## 依赖与版本
- Go 版本：`go 1.24`（来自 `go.mod`）
- 主要依赖：Gin、Kratos、Zap、GORM、Kafka、OpenTelemetry 等

## License
未在仓库中声明许可证。
