# goxf

`goxf` 是一个基于 Go 语言的高性能微服务框架，集成了服务治理、配置管理、日志、认证、Web 框架、gRPC、WebSocket 等常用能力，适合企业级微服务应用开发。

## 主要特性

- **模块化设计**：支持 server、client、auth、conf、log、util、i18n、hooks、flag 等模块，结构清晰，易于扩展。
- **多协议支持**：内置 HTTP（基于 Gin）、gRPC、WebSocket 等多种服务协议。
- **配置热加载**：支持配置文件热重载，配置变更自动生效。
- **优雅启停**：支持优雅启动与关闭，保证服务平滑退出。
- **服务注册与发现**：预留服务注册与发现接口，便于集成 Consul、Etcd 等注册中心。
- **高性能日志**：集成 zap 日志库，支持异步日志、日志切割。
- **国际化支持**：内置 i18n 国际化能力。
- **钩子机制**：支持生命周期钩子，方便扩展和自定义行为。
- **性能监控**：支持 pprof、metrics 等性能监控能力。
- **后台任务引擎**：内置高性能后台任务管理引擎 `job/cron` (定时任务) 与 `job/queue` (异步消息队列与高可用延迟任务)，利用泛型实现零侵入、强类型投递。

## 目录结构

```
.
├── application/    # 运行时与应用管理
├── auth/           # 认证与鉴权
├── client/         # 客户端相关
├── conf/           # 配置管理
├── ecode/          # 错误码管理
├── examples/       # 示例代码
├── flag/           # 命令行参数
├── hooks/          # 生命周期钩子
├── i18n/           # 国际化
├── log/            # 日志系统
├── server/         # 服务端相关
├── util/           # 工具包
├── config.go       # 配置结构定义
├── go.mod          # Go 依赖管理
├── go.sum          # Go 依赖校验
├── goxf.go         # 框架主入口
├── hlep.go         # 帮助信息
└── README.md       # 项目说明
```

## 快速开始

1. **安装依赖**

   ```bash
   go mod tidy
   ```

2. **编写配置文件**

   在项目根目录下创建 `config.yaml`，参考 `examples/config.yaml`。

3. **启动服务**

   ```go
   package main

   import (
       "github.com/xslasd/goxf"
   )

   func main() {
       service := goxf.NewService()
       // 添加自定义 server
       // service.AddServer(...)
       service.Run()
   }
   ```

4. **运行项目**

   ```bash
   go run main.go
   ```

## 主要依赖

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - Web 框架
- [go.uber.org/zap](https://github.com/uber-go/zap) - 日志库
- [xorm.io/xorm](https://xorm.io/) - ORM 框架
- [google.golang.org/grpc](https://grpc.io/) - gRPC
- [github.com/gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket
- [github.com/dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go) - JWT 认证

## 贡献

欢迎提交 issue 和 PR，完善文档和功能。

## 致谢 (Acknowledgments)

本项目在开发过程中，借鉴了优秀的开源微服务框架 [douyu/jupiter](https://github.com/douyu/jupiter) 的部分架构设计与底层抽象机制（如生命周期/配置监听等）。特此向开源社区及 Jupiter 团队的贡献致谢。

## License

[MIT](LICENSE) 