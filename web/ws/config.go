package ws

// Config WebSocket 配置
type Config struct {
	ReadBufferSize    int64 `toml:"readBufferSize"`    // 读缓冲区大小（字节）
	WriteBufferSize   int64 `toml:"writeBufferSize"`   // 写缓冲区大小（字节）
	MaxMessageSize    int64 `toml:"maxMessageSize"`    // 消息最大大小（字节）
	PingInterval      int   `toml:"pingInterval"`      // 心跳间隔（秒）
	PongTimeout       int   `toml:"pongTimeout"`       // Pong 超时时间（秒）
	EnableCompression bool  `toml:"enableCompression"` // 是否启用压缩
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		MaxMessageSize:    512 * 1024, // 512KB
		PingInterval:      30,         // 30秒
		PongTimeout:       60,         // 60秒
		EnableCompression: false,
	}
}
