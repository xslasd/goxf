package sse

import (
	"context"
	"errors"
	"sync"

	"github.com/xslasd/goxf/ecode"
)

// Event 表示发送给客户端的 SSE 事件
type Event struct {
	ID    string // 可选，事件 ID
	Event string // 可选，事件名称（默认是 message）
	Data  []byte // 必填，事件数据内容
}

// message 表示 Hub 内部的转发或广播包裹
type message struct {
	sender   string // 发送者客户端 ID
	receiver string // 接收者客户端 ID（如果是广播则为空）
	event    *Event // 事件体
}

// Client 表示一个单独的 SSE 客户端连接
type Client struct {
	ctx  context.Context // 关联的 context
	done chan struct{}   // 用于通知连接断开或关闭的通道
	id   string          // 客户端唯一标识符
	hub  *Hub            // 关联的 Hub 中心
	send chan *Event     // 消息投递缓冲区
	mu   sync.Mutex      // 互斥锁，用于保护 id 和关闭操作
	once sync.Once       // 确保 Close 操作只执行一次
}

// Context 返回客户端的 context
func (c *Client) Context() context.Context {
	return c.ctx
}

// Forward 将消息转发给指定的客户端（不包含事件名称）
func (c *Client) Forward(id string, data []byte) {
	c.ForwardEvent(id, "", data)
}

// ForwardEvent 将指定事件名称的消息转发给指定客户端
func (c *Client) ForwardEvent(id string, event string, data []byte) {
	c.mu.Lock()
	senderID := c.id
	c.mu.Unlock()
	msg := message{
		sender:   senderID,
		receiver: id,
		event: &Event{
			Event: event,
			Data:  data,
		},
	}
	select {
	case c.hub.forward <- msg:
	case <-c.hub.done:
	case <-c.done:
	}
}

// Broadcast 广播消息给所有其他客户端（不包含事件名称）
func (c *Client) Broadcast(data []byte) {
	c.BroadcastEvent("", data)
}

// BroadcastEvent 广播指定事件名称的消息给所有其他客户端
func (c *Client) BroadcastEvent(event string, data []byte) {
	c.mu.Lock()
	senderID := c.id
	c.mu.Unlock()
	msg := message{
		sender: senderID,
		event: &Event{
			Event: event,
			Data:  data,
		},
	}
	select {
	case c.hub.broadcast <- msg:
	case <-c.hub.done:
	case <-c.done:
	}
}

// SetID 修改客户端标识，更新 Hub 里的映射关系，并主动踢掉已存在的同名连接
func (c *Client) SetID(id string) {
	if id == "" {
		return
	}
	c.hub.Lock()
	defer c.hub.Unlock()

	c.mu.Lock()
	oldID := c.id
	if oldID == id {
		c.mu.Unlock()
		return
	}
	c.id = id
	c.mu.Unlock()

	// 删除旧映射
	delete(c.hub.clients, oldID)

	// 如果新 ID 已经存在连接，则关闭旧的连接
	if old, ok := c.hub.clients[id]; ok {
		old.Close()
	}

	c.hub.clients[id] = c
}

// GetID 获取客户端的唯一标识
func (c *Client) GetID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.id
}

// Send 向当前客户端发送数据（默认 message 事件）
func (c *Client) Send(message []byte) error {
	return c.SendEvent("", message)
}

// SendEvent 向当前客户端发送带有事件名称的数据
func (c *Client) SendEvent(event string, data []byte) error {
	ev := &Event{
		Event: event,
		Data:  data,
	}
	select {
	case c.send <- ev:
		return nil
	case <-c.done:
		return errors.New("sse connection closed")
	default:
		// 缓冲区满，说明客户端读取过慢，抛出缓冲区已满错误
		return ecode.SSEBufferFull
	}
}

// Close 关闭客户端通道，通知底层的推送循环退出
func (c *Client) Close() error {
	var err error
	c.once.Do(func() {
		close(c.done)
	})
	return err
}
