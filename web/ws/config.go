package ws

// Config WebSocket 配置
type Config struct {
	ReadBufferSize    int64    `toml:"readBufferSize"`
	WriteBufferSize   int64    `toml:"writeBufferSize"`
	MaxMessageSize    int64    `toml:"maxMessageSize"`
	PingInterval      int      `toml:"pingInterval"`
	PongTimeout       int      `toml:"pongTimeout"`
	EnableCompression bool     `toml:"enableCompression"`
	AllowedOrigins    []string `toml:"allowedOrigins"`
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
