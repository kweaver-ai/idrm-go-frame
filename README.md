# idrm-go-frame

[中文说明](README.zh.md)

A Go foundation framework for backend services, with app lifecycle, config, logging, transport, registry, storage, and utility modules.

## Overview
`idrm-go-frame` focuses on reusable building blocks for services: lifecycle management, configuration, logging, HTTP/Kafka transport, registry/discovery, storage helpers, and common utilities.

## Features
- Application lifecycle (start/stop, graceful shutdown, service registration).
- Configuration (multi-source, multi-format, hot reload).
- Logging (zap-based, multi-output, multi-level).
- Transport: REST (Gin) and Kafka (kafkax / kq).
- Telemetry wrappers for logs and tracing.
- Storage/cache utilities: gormx, redis tools, distributed locks.
- Common utilities: enum, encoding, syncx, utils.

## Layout
- `app.go`: application lifecycle and runner.
- `core/config`: configuration loaders and parsers.
- `core/logx`: logging infrastructure.
- `core/transport/rest`: REST server wrapper (Gin).
- `core/transport/mq/kafkax`: Kafka consumer/producer wrappers.
- `core/telemetry`: logging and tracing components.
- `core/store` / `core/redis_tool`: storage/cache helpers.
- `docs/`: component docs and examples.

## Quick Start
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

## Docs
- Config: `docs/component/config.md`
- Logging: `docs/component/logger.md`
- Enum: `docs/component/enum.md`
- Kafka: `docs/transport/kafkax.md`

## Requirements
- Go version: `go 1.24` (from `go.mod`)
- Key deps: Gin, Kratos, Zap, GORM, Kafka, OpenTelemetry, etc.

## License
No license is specified in this repository.
