# sgin (Gin Server Component)

`sgin` 是 `goxf` 框架对 [gin-gonic/gin](https://github.com/gin-gonic/gin) 的深度封装，为微服务提供了一个开箱即用、高度集成的高性能 HTTP 容器组件。它天然整合了框架配置体系、统一日志、国际化响应（I18n）以及跨域、JWT鉴权等常用中间件。

## 特性

*   **配置化启动**：支持从统一个配置中心解析 IP 绑定、慢查询阈值等参数。
*   **按需可插拔中间件**：
    *   自带 `Logger` 日志采集功能，区分终端控制台高亮调试模式及服务器生产模式。
    *   内置慢查询告警（Slow Query）捕获机制。
    *   灵活的跨域配置 (`CORS`) 注入。
    *   开箱即用的 JSON Web Token（JWT）鉴权中间件。
*   **I18n 及响应规范**：内置请求返回标准结构体（`Code`, `Msg`, `Data`），一键对接框架的 I18n 多语言环境拦截转译机制。
*   **服务生命周期管理**：完全兼容 `goxf` 顶层 `application` 平滑关机、统一启动等标准生命周期。

---

## 快速使用

### 0. 前置配置 (YAML)

在 `goxf` 统管的 `application.yaml` 或 `.toml` 等配置文件中增加如下节点，作为快速启用的前置条件：

```yaml
server:
  gin:
    default:
      Addr: "0.0.0.0:8080"                  # 绑定监听的地址及端口
      SlowQueryThresholdInMilli: 500        # 大于 500ms 的请求会被标记为 slow request
      AllowedOrigins:                       # CORS的跨域白名单（空则开放跨域限制）
        - "https://www.example.com"
```

### 1. 启动一个标准 HTTP 服务

你可以直接通过 `sgin.NewGinServer()` 构建一个由配置中心驱动的，具备自动日志上报、通用跨域功能的 HTTP 服务器。

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/xslasd/goxf/server/sgin"
)

func main() {
    // 构建 Gin 服务器，默认直接使用 config 中的配置参数
    server, err := sgin.NewGinServer()
    if err != nil {
        panic(err)
    }

    // 注册业务路由
    server.Engine.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    // (通常交给 goxf.Service 统一启动，单测/独立模块也可以单独调用原生的 server.server.ListenAndServe)
}
```

### 2. 统一的数据响应结构与 I18n 错误翻译

使用封装好的 `sgin.DefaultJSONWithI18nRes` 能极大简化 API 返回格式和多语言翻译的编写成本：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/xslasd/goxf/server/sgin"
    "github.com/xslasd/goxf/ecode"
)

func UserInfo(c *gin.Context) {
    var user struct{ Name string }
    err := getUser(&user)
    
    if err != nil {
        // ecode.UserNotFound 会被底层中间件自动翻译为目标语言（需配合框架 I18n 使用）
        sgin.DefaultJSONWithI18nRes(c, nil, ecode.UserNotFound)
        return
    }

    // 正常返回数据，err=nil 表示状态码=0 成功
    sgin.DefaultJSONWithI18nRes(c, user, nil)
}
```

响应 JSON 样例：

```json
{
    "code": 0,
    "msg": "Success",
    "data": {
        "Name": "goxf_user"
    }
}
```

### 3. JWT 鉴权中间件保护路由

可以直接通过 `sgin.JWTAuthMiddleware` 为私密服务做请求拦截保护。

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/xslasd/goxf/server/sgin"
)

func verifyTokenFunc(c *gin.Context, token string) error {
    // 此处可调用你自己的业务 token 解码验证逻辑
    if token == "Valid_Token" {
        return nil
    }
    // 返回框架级错误规范会自动化转换为 JSON
    return ecode.Unauthorized
}

func setupRouter(engine *gin.Engine) {
    // 注入 JWT 鉴权中间件
    authGroup := engine.Group("/api/v1")
    authGroup.Use(sgin.JWTAuthMiddleware(verifyTokenFunc))
    
    authGroup.GET("/profile", func(c *gin.Context) {
        c.JSON(200, gin.H{"msg": "Authorized!"})
    })
}
```

### 4. 慢查询与大延迟请求挂载告警 (TimeoutEvent)

`sgin` 框架内置了对请求耗时的统一收集功能。如果在 YAML 配置中设定了 `SlowQueryThresholdInMilli` 且接口处理时长超过了该阈值（默认 `500ms`），除了在终端打出 Slow Log 以外，还会触发 `TimeoutEventFunc` 钩子函数。你可以利用它实现系统级警告推送（如钉钉或飞书）、特定慢查询专项数据上报等：

```go
import (
    "fmt"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/xslasd/goxf/server/sgin"
)

// 定义你的慢查询回调处理逻辑
func myTimeoutHandler(c *gin.Context, route string, cost time.Duration) {
    clientIP := c.ClientIP()
    userAgent := c.GetHeader("User-Agent")
    
    alertMsg := fmt.Sprintf("[告警] 发现慢请求！路由: %s, 耗时: %v, IP: %s, UA: %s", route, cost, clientIP, userAgent)
    fmt.Println(alertMsg)
    // 具体推送可以对接到您所在的告警中心...
}

func main() {
    // 构建 Server 时注入我们提前写好的超时监控 Event
    server, err := sgin.NewGinServer(
        sgin.WithTimeoutEvent(myTimeoutHandler),
    )
    if err != nil {
        panic(err)
    }

    server.Engine.GET("/heavy-task", func(c *gin.Context) {
        time.Sleep(600 * time.Millisecond) // 会触发大耗时告警
        c.JSON(200, gin.H{"msg": "done"})
    })
}
```

---

## 进阶配置选项 (Options)

在调用 `sgin.NewGinServer(opts ...Option)` 时可以通过以下函数增强、自定义你的服务器容器。

| Option 方法名 | 参数说明 | 功能描述 |
| --- | --- | --- |
| `WithConfPrefix(prefix string)` | 配置前缀层级 | 设置用于拉取配置文件中的字典级。默认为 `server.gin` |
| `WithConfName(name string)` | 配置名称 | 用于与 `Prefix` 拼接。默认为 `default`。拉取目标: `server.gin.default` |
| `WithMiddleware(handler)` | gin.HandlerFunc | 向容器最根部注入额外的中间件 (优先于路由执行) |
| `WithEnableConsole(bool)` | 是否开启彩色日志 | 设置为 `true` 会在控制台打印美化的请求日志 `debugMiddleware` |
| `WithEnableMetric(bool)` | 是否开启监控埋点 | 将请求吞吐延迟同步至 Prometheus 埋点 (待拓展) |
| `WithEnableTrace(bool)` | 是否开启拓扑追踪 | 将 HTTP 请求链接入 Jaeger 分布式追踪网络 (待拓展) |
| `WithCorsOptions(opts)` | cors.Options | 配置跨域行为，比如允许包含特定头部、特定 Origin |
| `WithTimeoutEvent(event)` | func(c, route, cost) | 配置当慢查询（由阈值判断）触发时的附加事件或报警回调 |
