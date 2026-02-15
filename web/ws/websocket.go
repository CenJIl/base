package ws

import (
	"net/http"
	"time"

	"github.com/CenJIl/base/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gorilla/websocket"
)

// Upgrader HTTP 升级为 WebSocket 的配置
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: 生产环境应该验证 Origin
		return true
	},
	HandshakeTimeout: 10 * time.Second,
}

// ConfigureUpgrader 根据配置升级 Upgrader
func ConfigureUpgrader(config Config) {
	Upgrader.ReadBufferSize = int(config.ReadBufferSize)
	Upgrader.WriteBufferSize = int(config.WriteBufferSize)
	Upgrader.EnableCompression = config.EnableCompression
}

// HandleWebSocketHTTP WebSocket HTTP 升级处理函数
//
// 使用 Hub 管理连接池
//
// 使用方式：
//
//	hub := ws.NewHub()
//	h.GET("/ws", func(ctx context.Context, c *app.RequestContext) {
//	    conn, err := ws.UpgradeHTTP(c)
//	    if err != nil {
//	        logger.Errorf("WS upgrade failed: %v", err)
//	        return
//	    }
//	    connection := ws.NewConnection(conn, hub)
//	    hub.Register(connection)
//	    go connection.ReadPump()
//	    go connection.WritePump()
//	})
func UpgradeHTTP(c *app.RequestContext) (*websocket.Conn, error) {
	// 注意：Hertz 和 gorilla/websocket 的接口不完全兼容
	// 实际使用中，建议直接使用 gorilla/websocket 的标准用法
	// 或者使用 Hertz 内置的 WebSocket 支持（如果有的话）
	logger.Errorf("[WS] Upgrade functionality requires manual implementation")
	return nil, http.ErrNotSupported
}
