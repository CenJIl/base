package ws

import (
	"sync"

	"github.com/CenJIl/base/logger"
)

// Hub WebSocket 连接池
//
// 管理所有 WebSocket 连接，支持广播和点对点消息
type Hub struct {
	connections map[string]*Connection // 连接映射（ID -> Connection）
	register    chan *Connection       // 注册连接
	unregister  chan *Connection       // 注销连接
	broadcast   chan []byte           // 广播消息
	mu          sync.RWMutex          // 读写锁
	onMessage  func(*Connection, []byte) // 消息处理回调
}

// NewHub 创建新的连接池
//
// 使用方式：
//   hub := ws.NewHub()
//   go hub.Run()
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		broadcast:   make(chan []byte, 256),
	}
}

// Run 启动连接池（阻塞运行）
//
// 使用方式：
//   hub := ws.NewHub()
//   go hub.Run()  // 在独立协程中运行
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn.ID()] = conn
			h.mu.Unlock()
			logger.Infof("[WS] Connection registered: %s (total: %d)", conn.ID(), len(h.connections))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.ID()]; ok {
				delete(h.connections, conn.ID())
				conn.Close()
			}
			h.mu.Unlock()
			logger.Infof("[WS] Connection unregistered: %s (total: %d)", conn.ID(), len(h.connections))

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, conn := range h.connections {
				select {
				case conn.send <- message:
					// 消息已发送
				default:
					// 发送队列已满，关闭连接
					logger.Warnf("[WS] Broadcast buffer full for connection: %s", conn.ID())
					h.unregister <- conn
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register 注册连接
//
// 使用方式：
//   hub.Register(conn)
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// Unregister 注销连接
//
// 使用方式：
//   hub.Unregister(conn)
func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

// Broadcast 广播消息给所有连接
//
// 使用方式：
//   hub.Broadcast([]byte("system notification"))
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendTo 发送消息给指定连接
//
// 使用方式：
//   hub.SendTo("conn-id", []byte("private message"))
func (h *Hub) SendTo(connID string, message []byte) error {
	h.mu.RLock()
	conn, ok := h.connections[connID]
	h.mu.RUnlock()

	if !ok {
		return ErrConnectionNotFound
	}

	conn.Send(message)
	return nil
}

// GetConnection 获取指定连接
//
// 使用方式：
//   conn := hub.GetConnection("conn-id")
func (h *Hub) GetConnection(connID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, ok := h.connections[connID]
	return conn, ok
}

// GetConnectionCount 获取当前连接数
//
// 使用方式：
//   count := hub.GetConnectionCount()
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// GetConnections 获取所有连接
//
// 使用方式：
//   conns := hub.GetConnections()
func (h *Hub) GetConnections() []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		conns = append(conns, conn)
	}
	return conns
}

// OnMessage 设置消息处理回调
//
// 使用方式：
//   hub.OnMessage(func(conn *ws.Connection, msg []byte) {
//       logger.Infof("Received: %s", msg)
//   })
func (h *Hub) OnMessage(handler func(*Connection, []byte)) {
	h.onMessage = handler
}

// OnMessage 内部消息处理（由 Connection 调用）
func (h *Hub) onMessageHandler(conn *Connection, message []byte) {
	if h.onMessage != nil {
		h.onMessage(conn, message)
	}
}

// 错误定义
var (
	ErrConnectionNotFound = &HubError{Code: 404, Message: "Connection not found"}
)

// HubError Hub 错误类型
type HubError struct {
	Code    int
	Message string
}

func (e *HubError) Error() string {
	return e.Message
}
