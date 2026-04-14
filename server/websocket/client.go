package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xslasd/goxf/ecode"
)

type message struct {
	sender   string
	receiver string
	msg      []byte
}

type Client struct {
	ctx  context.Context
	done chan struct{}
	id   string
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // 消息缓冲区
	mu   sync.Mutex
	once sync.Once
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) Forward(id string, data []byte) {
	c.mu.Lock()
	senderID := c.id
	c.mu.Unlock()
	msg := message{
		sender:   senderID,
		receiver: id,
		msg:      data,
	}
	select {
	case c.hub.forward <- msg:
	case <-c.hub.done:
	case <-c.done:
	}
}

func (c *Client) Broadcast(data []byte) {
	c.mu.Lock()
	senderID := c.id
	c.mu.Unlock()
	msg := message{
		sender: senderID,
		msg:    data,
	}
	select {
	case c.hub.broadcast <- msg:
	case <-c.hub.done:
	case <-c.done:
	}
}

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

	//删除旧映射
	delete(c.hub.clients, oldID)

	if old, ok := c.hub.clients[id]; ok {
		old.Close()
	}

	c.hub.clients[id] = c
}

func (c *Client) GetID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.id
}

func (c *Client) Send(message []byte) error {
	select {
	case c.send <- message:
		return nil
	case <-c.done:
		return websocket.ErrCloseSent
	default:
		// 缓冲区满，说明客户端处理太慢，建议返回错误或直接关闭慢连接
		return ecode.WSBufferFull
	}
}

// writePump 将消息从缓冲区推送到 websocket 连接，并负责发送心跳
func (c *Client) writePump() {
	ticker := time.NewTicker(c.hub.keepAliveTime)
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case msg := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.hub.writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.hub.cliErrHandler(c, err)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.hub.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.hub.cliErrHandler(c, err)
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) readStream() {
	defer func() {
		c.Close()
	}()
	c.conn.SetReadLimit(c.hub.maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(c.hub.readWait)); err != nil {
		c.hub.cliErrHandler(c, err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(c.hub.readWait))
	})
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.cliErrHandler(c, err)
			return
		}
		c.hub.msgHandler(c, msg)
	}
}

func (c *Client) Close() error {
	var err error
	c.once.Do(func() {
		err = c.conn.Close()
		close(c.done) // 关闭发送通道，通知 writePump 退出
	})
	return err
}
