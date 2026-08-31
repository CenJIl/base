package ws

import (
	"sync"
	"time"

	"github.com/CenJIl/base/logger"
	"github.com/gorilla/websocket"
)

// Connection WebSocket 连接封装
type Connection struct {
	hub            *Hub
	ws             *websocket.Conn
	send           chan []byte
	id             string
	maxMessageSize int64
	pongWait       time.Duration
	pingPeriod     time.Duration
	closeOnce      sync.Once
}

// NewConnection 创建新连接
//
// 使用方式：
//
//	conn := ws.NewConnection(wsConn, hub)
func NewConnection(wsConn *websocket.Conn, hub *Hub) *Connection {
	config := currentConnectionConfig()
	return &Connection{
		hub:            hub,
		ws:             wsConn,
		send:           make(chan []byte, 256),
		id:             generateConnID(),
		maxMessageSize: config.MaxMessageSize,
		pongWait:       time.Duration(config.PongTimeout) * time.Second,
		pingPeriod:     time.Duration(config.PingInterval) * time.Second,
	}
}

// ReadPump 读取协程
//
// 从 WebSocket 读取消息并广播到 Hub
func (c *Connection) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.ws.Close()
	}()

	c.ws.SetReadLimit(c.maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(c.pongWait))

	// 配置 Pong 处理器
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("[WS] Read error: %v", err)
			}
			break
		}

		// 处理接收到的消息
		c.hub.onMessageHandler(c, message)
	}
}

// WritePump 写入协程
//
// 从 send 队列读取消息并写入 WebSocket
func (c *Connection) WritePump() {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了连接
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Errorf("[WS] Write error: %v", err)
				return
			}

		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			// 发送 Ping
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send 发送消息（非阻塞）
//
// 使用方式：
//
//	conn.Send([]byte("hello"))
func (c *Connection) Send(message []byte) {
	select {
	case c.send <- message:
		// 消息已加入发送队列
	default:
		// 发送队列已满，关闭连接
		logger.Warnf("[WS] Send buffer full, closing connection: %s", c.id)
		c.hub.Unregister(c)
	}
}

// Close 关闭连接
//
// 使用方式：
//
//	conn.Close()
func (c *Connection) Close() {
	c.closeOnce.Do(func() { close(c.send) })
}

// ID 获取连接 ID
//
// 使用方式：
//
//	id := conn.ID()
func (c *Connection) ID() string {
	return c.id
}

const writeWait = 10 * time.Second

// generateConnID 生成连接 ID
func generateConnID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
