# SSE (Server-Sent Events) 长连接封装模块

`sse` 模块提供了基于 HTTP 协议的轻量级单向推送长连接封装。其接口和设计模式完全参考了框架原有的 `server/websocket` 包，为业务提供了几乎一致的使用体验。

## 核心特性

* **全自动生命周期维护**：只需调用 `hub.Upgrade` 即可完成 HTTP 协议握手，剩下的资源回收、心跳保活、连接移除全部由底层框架自动托管。
* **支持多种投递模式**：
  * 点对点数据转发（`Forward` / `ForwardEvent`）
  * 全局广播（`Broadcast` / `BroadcastEvent`）
* **心跳自动保活 (Keep-Alive)**：内置定时器定时发送 `: keepalive` 注释行，能够完美刷新 Nginx 等中间代理或防火墙的空闲超时时间，同时不会干扰客户端的 `onmessage` 逻辑。
* **抢占踢人机制**：当同一个客户端 ID 在另一处重新发起连接时，老连接会自动被强制断开（Kickout 机制）。
* **多行数据完美兼容**：即使推送的数据包含换行符（如 JSON 格式），内部会自动解析并按行增加 `data:` 前缀，符合标准 SSE 规范。

---

## 核心结构体与选项

### 1. Hub (控制器中心)

```go
// 创建一个 Hub
hub, err := sse.NewHub(
    sse.WithKeepAliveTime(15 * time.Second), // 设置心跳间隔，默认 10s
    sse.WithCliCloseHandler(func(c *sse.Client) {
        fmt.Printf("客户端断开连接: %s\n", c.GetID())
    }),
    sse.WithCliErrHandler(func(c *sse.Client, err error) {
        fmt.Printf("客户端写入出错: %s, 错误: %v\n", c.GetID(), err)
    }),
)
```

### 2. Client (客户端连接)

* `Context() context.Context`：获取连接绑定的 context，以便感知退出信号。
* `GetID() string`：获取客户端 ID。
* `SetID(id string)`：重新设置 ID，会自动维护 Hub 里的在线映射。
* `Send(data []byte) error`：发送纯数据（默认 message 事件）。
* `SendEvent(event string, data []byte) error`：发送特定事件名称的数据。
* `Forward(receiverId string, data []byte)`：由当前客户端向特定客户端转发消息。
* `Broadcast(data []byte)`：由当前客户端向所有人（除了自己）广播消息。
* `Close() error`：主动断开连接。

---

## 服务端使用示例 (基于 Gin)

由于 `Upgrade` 的入参为 `http.ResponseWriter` 和 `*http.Request`，在 Gin 中极易进行集成：

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xslasd/goxf/server/sse"
)

func main() {
	// 1. 初始化 Hub
	hub, err := sse.NewHub(
		sse.WithKeepAliveTime(10 * time.Second),
		sse.WithCliCloseHandler(func(c *sse.Client) {
			log.Printf("客户端 %s 离开了", c.GetID())
		}),
	)
	if err != nil {
		log.Fatalf("初始化 Hub 失败: %v", err)
	}
	defer hub.Close()

	// 2. 挂载 Gin 路由
	r := gin.Default()
	r.GET("/sse", func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id 不能为空"})
			return
		}

		// 升级为 SSE 长连接，该方法会阻塞，直到连接由于客户端断开、服务端踢人或 Hub 关闭而退出
		err := hub.Upgrade(c.Request.Context(), c.Writer, c.Request, id, 
			sse.WithStartHandler(func(cli *sse.Client) {
				log.Printf("客户端 %s 连接成功", cli.GetID())
				
				// 立即推送一条欢迎消息
				_ = cli.SendEvent("welcome", []byte("连接建立成功！"))
			}),
		)
		if err != nil {
			log.Printf("连接异常: %v", err)
		}
	})

	// 3. 其他路由：用于触发数据推送
	r.POST("/push", func(c *gin.Context) {
		targetID := c.PostForm("target_id")
		content := c.PostForm("content")
		
		// 通过 Hub 点对点推送到指定客户端
		hub.ForwardEvent(targetID, "notification", []byte(content))
		
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/broadcast", func(c *gin.Context) {
		content := c.PostForm("content")
		
		// 通过 Hub 广播到所有客户端
		hub.BroadcastEvent("announcement", []byte(content))
		
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}
```

---

## 客户端/前端使用示例 (JavaScript)

在前端，您可以使用标准的 HTML5 `EventSource` 轻松对接：

```javascript
// 1. 建立 SSE 连接并传递唯一的客户端 ID
const eventSource = new EventSource('http://localhost:8080/sse?id=user_12345');

// 2. 监听默认事件 (通过 client.Send 发送的消息)
eventSource.onmessage = function(event) {
    console.log("收到默认消息:", event.data);
};

// 3. 监听自定义事件名称消息 (通过 client.SendEvent 发送的消息)
eventSource.addEventListener('welcome', function(event) {
    console.log("欢迎语:", event.data);
});

eventSource.addEventListener('notification', function(event) {
    console.log("收到新通知:", event.data);
});

eventSource.addEventListener('announcement', function(event) {
    console.log("收到全员公告:", event.data);
});

// 4. 监听连接建立
eventSource.onopen = function() {
    console.log("SSE 连接已建立");
};

// 5. 监听错误 / 自动重连
eventSource.onerror = function(err) {
    console.error("SSE 连接出错或被断开", err);
};
```
