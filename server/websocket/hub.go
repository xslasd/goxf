package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xslasd/goxf/ecode"
)

const (
	Forward = iota + 1
	Broadcast
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10024,
	WriteBufferSize: 10024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	sync.RWMutex
	clients   map[string]*Client
	forward   chan message
	broadcast chan message

	done chan struct{}

	hubErrHandler   HubErrorHandler
	cliErrHandler   CliErrHandler
	cliCloseHandler CliCloseHandler
	msgHandler      MessageHandler
	upGrader        websocket.Upgrader

	readWait       time.Duration
	writeWait      time.Duration
	keepAliveTime  time.Duration
	maxMessageSize int64
}

type CliCloseHandler func(c *Client)

type HubErrorHandler func(int, message, error)

type CliErrHandler func(*Client, error)

type MessageHandler func(c *Client, msg []byte)

type Option func(*Hub)

func WithKeepAliveTime(t time.Duration) Option {
	return func(hub *Hub) {
		hub.keepAliveTime = t
	}
}

func WithReadWait(t time.Duration) Option {
	return func(hub *Hub) {
		hub.readWait = t
	}
}

func WithWriteWait(t time.Duration) Option {
	return func(hub *Hub) {
		hub.writeWait = t
	}
}

func WithMaxMessageSize(size int64) Option {
	return func(hub *Hub) {
		hub.maxMessageSize = size
	}
}

func WithCliCloseHandler(handler CliCloseHandler) Option {
	return func(hub *Hub) {
		hub.cliCloseHandler = handler
	}
}

func WithCliErrHandler(handler CliErrHandler) Option {
	return func(hub *Hub) {
		hub.cliErrHandler = handler
	}
}

func WithHubErrorHandler(handler HubErrorHandler) Option {
	return func(hub *Hub) {
		hub.hubErrHandler = handler
	}
}

func WithMessageHandler(handler MessageHandler) Option {
	return func(hub *Hub) {
		hub.msgHandler = handler
	}
}

func WithUpGrader(upGrader websocket.Upgrader) Option {
	return func(hub *Hub) {
		hub.upGrader = upGrader
	}
}

func NewHub(opts ...Option) (*Hub, error) {
	h := &Hub{
		clients:   make(map[string]*Client),
		forward:   make(chan message, 100),
		broadcast: make(chan message, 100),
		done:      make(chan struct{}),

		readWait:       20 * time.Second,
		writeWait:      10 * time.Second,
		keepAliveTime:  10 * time.Second,
		maxMessageSize: 10240,
		upGrader:       upgrader,
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
	if h.msgHandler == nil {
		return nil, ecode.WSMessageHandlerIsNil
	}
	go h.run()
	return h, nil
}

func (h *Hub) Close() error {
	select {
	case <-h.done:
		return nil
	default:
		close(h.done)
	}
	return nil
}

func (h *Hub) Forward(id string, data []byte) {
	msg := message{
		receiver: id,
		msg:      data,
	}
	select {
	case h.forward <- msg:
	case <-h.done:
	}
}

func (h *Hub) Broadcast(data []byte) {
	msg := message{
		msg: data,
	}
	select {
	case h.broadcast <- msg:
	case <-h.done:
	}
}

func (h *Hub) OnlineClientCount() int {
	h.RLock()
	defer h.RUnlock()
	return len(h.clients)
}

func (h *Hub) IsClientOnline(id string) bool {
	h.RLock()
	defer h.RUnlock()
	_, ok := h.clients[id]
	return ok
}

func (h *Hub) GetClient(id string) (*Client, bool) {
	h.RLock()
	defer h.RUnlock()
	c, ok := h.clients[id]
	return c, ok
}

func (h *Hub) run() {
	defer func() {
		h.RLock()
		// 1. 先拷贝出来，避免在持有锁的情况下调用可能触发锁竞争的操作
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
					continue
				}
				// 直接向 client.send 投递，不再阻塞 Hub 循环
				_ = c.Send(v.msg)
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
		return ecode.WSClientNotExist
	}
	return re.Send(msg.msg)
}

type upgradeConfig struct {
	header       http.Header
	startHandler StartHandler
}
type StartHandler func(cli *Client)

type UpgradeOptions func(*upgradeConfig)

func WithHeader(header http.Header) UpgradeOptions {
	return func(cfg *upgradeConfig) {
		cfg.header = header
	}
}
func WithStartHandler(event StartHandler) UpgradeOptions {
	return func(cfg *upgradeConfig) {
		cfg.startHandler = event
	}
}

func (h *Hub) Upgrade(ctx context.Context, w http.ResponseWriter, r *http.Request, cliId string, opts ...UpgradeOptions) error {
	if cliId == "" {
		return ecode.WSClientIdIsNull
	}
	cfg := new(upgradeConfig)
	for _, o := range opts {
		o(cfg)
	}
	conn, err := h.upGrader.Upgrade(w, r, cfg.header)
	if err != nil {
		return err
	}
	client := &Client{
		ctx:  ctx,
		id:   cliId,
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	h.Lock()
	if old, ok := h.clients[cliId]; ok {
		old.Close() // 主动关闭旧的重名连接
	}
	h.clients[client.id] = client
	h.Unlock()

	go client.readStream()
	go client.writePump()
	if cfg.startHandler != nil {
		go cfg.startHandler(client)
	}

	<-client.done

	currID := client.GetID()
	h.Lock()
	if cli, ok := h.clients[currID]; ok && cli == client {
		delete(h.clients, currID)
	}
	h.Unlock()
	h.cliCloseHandler(client)
	return nil
}
