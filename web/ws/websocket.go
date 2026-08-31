package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/gorilla/websocket"
)

// Upgrader uses gorilla/websocket through Hertz's official http.Handler adaptor.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(*http.Request) bool {
		return false
	},
	HandshakeTimeout: 10 * time.Second,
}

var (
	upgraderMu       sync.RWMutex
	connectionConfig = DefaultConfig()
)

// ConfigureUpgrader applies buffer, compression, and explicit origin settings.
func ConfigureUpgrader(config Config) {
	if config.ReadBufferSize <= 0 {
		config.ReadBufferSize = 1024
	}
	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = 1024
	}
	if config.MaxMessageSize <= 0 {
		config.MaxMessageSize = 512 * 1024
	}
	if config.PingInterval <= 0 {
		config.PingInterval = 30
	}
	if config.PongTimeout <= config.PingInterval {
		config.PongTimeout = config.PingInterval * 2
	}
	allowed := make(map[string]struct{}, len(config.AllowedOrigins))
	for _, origin := range config.AllowedOrigins {
		allowed[origin] = struct{}{}
	}

	upgraderMu.Lock()
	connectionConfig = config
	Upgrader.ReadBufferSize = int(config.ReadBufferSize)
	Upgrader.WriteBufferSize = int(config.WriteBufferSize)
	Upgrader.EnableCompression = config.EnableCompression
	Upgrader.CheckOrigin = func(r *http.Request) bool {
		if len(allowed) == 0 {
			return false
		}
		_, ok := allowed[r.Header.Get("Origin")]
		return ok
	}
	upgraderMu.Unlock()
}

func currentConnectionConfig() Config {
	upgraderMu.RLock()
	defer upgraderMu.RUnlock()
	return connectionConfig
}

// UpgradeHTTP adapts a Gorilla WebSocket endpoint for Hertz.
func UpgradeHTTP(handler func(*websocket.Conn)) app.HandlerFunc {
	return adaptor.HertzHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler == nil {
			http.Error(w, "websocket handler is required", http.StatusInternalServerError)
			return
		}
		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		handler(conn)
	}))
}
