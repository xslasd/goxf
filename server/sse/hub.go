package sse

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/xslasd/goxf/ecode"
)

const (
	Forward = iota + 1
	Broadcast
)

// Hub 管理所有在线的 SSE 客户端连接以及消息路由
type Hub struct {
	sync.RWMutex
	clients   map[string]*Client // 在线客户端映射表
	forward   chan message       // 转发消息队列
	broadcast chan message       // 广播消息队列

	done chan struct{} // 标记 Hub 是否关闭的通道

	hubErrHandler   HubErrorHandler   // 路由错误处理器
	cliErrHandler   CliErrHandler     // 客户端写入错误处理器
	cliCloseHandler CliCloseHandler   // 客户端连接关闭回调

	keepAliveTime  time.Duration // 心跳包（注释行）发送周期
}

// CliCloseHandler 客户端连接关闭回调函数类型
type CliCloseHandler func(c *Client)

// HubErrorHandler Hub 路由处理出现错误时的回调函数类型
type HubErrorHandler func(action int, msg message, err error)

// CliErrHandler 客户端发送/写入网络数据出错时的回调函数类型
type CliErrHandler func(c *Client, err error)

// Option 定义 Hub 的参数配置选项
type Option func(*Hub)

// WithKeepAliveTime 设置心跳包（注释行）发送周期
func WithKeepAliveTime(t time.Duration) Option {
	return func(hub *Hub) {
		hub.keepAliveTime = t
	}
}

// WithCliCloseHandler 设置客户端连接关闭回调
func WithCliCloseHandler(handler CliCloseHandler) Option {
	return func(hub *Hub) {
		hub.cliCloseHandler = handler
	}
}

// WithCliErrHandler 设置客户端写入出错回调
func WithCliErrHandler(handler CliErrHandler) Option {
	return func(hub *Hub) {
		hub.cliErrHandler = handler
	}
}

// WithHubErrorHandler 设置 Hub 内部路由投递出错的回调
func WithHubErrorHandler(handler HubErrorHandler) Option {
	return func(hub *Hub) {
		hub.hubErrHandler = handler
	}
}

// NewHub 创建并启动一个新的 SSE Hub
func NewHub(opts ...Option) (*Hub, error) {
	h := &Hub{
		clients:   make(map[string]*Client),
		forward:   make(chan message, 100),
		broadcast: make(chan message, 100),
		done:      make(chan struct{}),

		keepAliveTime:  10 * time.Second,
		cliErrHandler: func(client *Client, err error) {
			return
		},
		cliCloseHandler: func(c *Client) {
			return
		},
		hubErrHandler: func(i int, m message, err error) {
			return
		},
	}
	for _, o := range opts {
		o(h)
	}
	go h.run()
	return h, nil
}

// Close 优雅关闭 Hub 并断开所有在线连接
func (h *Hub) Close() error {
	select {
	case <-h.done:
		return nil
	default:
		close(h.done)
	}
	return nil
}

// Forward 转发纯数据消息给特定客户端
func (h *Hub) Forward(id string, data []byte) {
	h.ForwardEvent(id, "", data)
}

// ForwardEvent 转发指定事件名称的消息给特定客户端
func (h *Hub) ForwardEvent(id string, event string, data []byte) {
	msg := message{
		receiver: id,
		event: &Event{
			Event: event,
			Data:  data,
		},
	}
	select {
	case h.forward <- msg:
	case <-h.done:
	}
}

// Broadcast 广播纯数据消息给所有在线客户端
func (h *Hub) Broadcast(data []byte) {
	h.BroadcastEvent("", data)
}

// BroadcastEvent 广播指定事件名称的消息给所有在线客户端
func (h *Hub) BroadcastEvent(event string, data []byte) {
	msg := message{
		event: &Event{
			Event: event,
			Data:  data,
		},
	}
	select {
	case h.broadcast <- msg:
	case <-h.done:
	}
}

// OnlineClientCount 获取当前在线客户端总数
func (h *Hub) OnlineClientCount() int {
	h.RLock()
	defer h.RUnlock()
	return len(h.clients)
}

// IsClientOnline 判断指定客户端是否在线
func (h *Hub) IsClientOnline(id string) bool {
	h.RLock()
	defer h.RUnlock()
	_, ok := h.clients[id]
	return ok
}

// GetClient 获取在线客户端连接实例
func (h *Hub) GetClient(id string) (*Client, bool) {
	h.RLock()
	defer h.RUnlock()
	c, ok := h.clients[id]
	return c, ok
}

// run 在后台协程中运行，负责分发 forward 和 broadcast 消息
func (h *Hub) run() {
	defer func() {
		h.RLock()
		clients := make([]*Client, 0, len(h.clients))
		for _, c := range h.clients {
			clients = append(clients, c)
		}
		h.RUnlock()
		for _, c := range clients {
			c.Close()
		}
	}()
	for {
		select {
		case v, ok := <-h.forward:
			if !ok {
				return
			}
			err := h.writeHandler(v)
			if err != nil {
				h.hubErrHandler(Forward, v, err)
			}
		case v, ok := <-h.broadcast:
			if !ok {
				return
			}
			h.RLock()
			for _, c := range h.clients {
				if c.GetID() == v.sender {
					// 不广播给自己
					continue
				}
				_ = c.SendEvent(v.event.Event, v.event.Data)
			}
			h.RUnlock()
		case <-h.done:
			return
		}
	}
}

func (h *Hub) writeHandler(msg message) error {
	h.RLock()
	re, ok := h.clients[msg.receiver]
	h.RUnlock()
	if !ok {
		return ecode.SSEClientNotExist
	}
	return re.SendEvent(msg.event.Event, msg.event.Data)
}

// upgradeConfig 包含长连接握手/处理阶段的可选配置项
type upgradeConfig struct {
	header       http.Header
	startHandler StartHandler
}

// StartHandler 客户端就绪后的回调函数类型
type StartHandler func(cli *Client)

// UpgradeOptions 定义 Upgrade 阶段的配置函数
type UpgradeOptions func(*upgradeConfig)

// WithHeader 设置连接初始化响应的额外 Header
func WithHeader(header http.Header) UpgradeOptions {
	return func(cfg *upgradeConfig) {
		cfg.header = header
	}
}

// WithStartHandler 设置客户端连接就绪后自动触发的事件回调
func WithStartHandler(event StartHandler) UpgradeOptions {
	return func(cfg *upgradeConfig) {
		cfg.startHandler = event
	}
}

// Upgrade 将普通的 HTTP 连接升级为 SSE 长连接响应流，并阻塞运行直到连接断开
func (h *Hub) Upgrade(ctx context.Context, w http.ResponseWriter, r *http.Request, cliId string, opts ...UpgradeOptions) error {
	if cliId == "" {
		return ecode.SSEClientIdIsNull
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	cfg := new(upgradeConfig)
	for _, o := range opts {
		o(cfg)
	}

	// 设置 SSE 协议规范必需的 Response Header
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // 避免 Nginx 缓存代理

	// 写入自定义 Header
	for k, v := range cfg.header {
		w.Header()[k] = v
	}

	// 立即刷新标头到客户端建立握手
	flusher.Flush()

	client := &Client{
		ctx:  ctx,
		id:   cliId,
		hub:  h,
		send: make(chan *Event, 256),
		done: make(chan struct{}),
	}

	h.Lock()
	if old, ok := h.clients[cliId]; ok {
		old.Close() // 踢掉同名客户端
	}
	h.clients[client.id] = client
	h.Unlock()

	// 触发长连接就绪回调
	if cfg.startHandler != nil {
		go cfg.startHandler(client)
	}

	ticker := time.NewTicker(h.keepAliveTime)
	defer func() {
		ticker.Stop()
		client.Close()

		currID := client.GetID()
		h.Lock()
		if cli, ok := h.clients[currID]; ok && cli == client {
			delete(h.clients, currID)
		}
		h.Unlock()
		h.cliCloseHandler(client)
	}()

	for {
		select {
		case ev := <-client.send:
			_, err := w.Write(ev.Marshal())
			if err != nil {
				h.cliErrHandler(client, err)
				return err
			}
			flusher.Flush()
		case <-ticker.C:
			// 发送注释行作为 KeepAlive 心跳，防止代理端超时断开
			_, err := w.Write([]byte(": keepalive\n\n"))
			if err != nil {
				h.cliErrHandler(client, err)
				return err
			}
			flusher.Flush()
		case <-r.Context().Done():
			// 客户端主动断开
			return nil
		case <-client.done:
			// 服务端主动关闭该 Client
			return nil
		case <-h.done:
			// Hub 被关闭
			return nil
		}
	}
}

// Marshal 将 Event 格式化为标准的 SSE 纯文本数据块
func (e *Event) Marshal() []byte {
	var buf bytes.Buffer
	if e.ID != "" {
		buf.WriteString("id: ")
		buf.WriteString(e.ID)
		buf.WriteByte('\n')
	}
	if e.Event != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Event)
		buf.WriteByte('\n')
	}
	if len(e.Data) > 0 {
		// 按照换行符拆分，对每一行添加 "data: " 前缀
		lines := bytes.Split(e.Data, []byte("\n"))
		for _, line := range lines {
			buf.WriteString("data: ")
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}
	buf.WriteByte('\n') // 必须以空行结尾，表示当前事件块的传输结束
	return buf.Bytes()
}
