package jwt

type Config struct {
	Secret      string   `toml:"secret"`      // JWT 密钥（必须配置）
	Realm       string   `toml:"realm"`       // 领域名，默认 "jwt"
	Timeout     int      `toml:"timeout"`     // 过期时间（秒），默认 3600（1小时）
	MaxRefresh  int      `toml:"maxRefresh"`  // 最大刷新时间（秒），默认 7200（2小时）
	IdentityKey string   `toml:"identityKey"` // 身份标识键，默认 "identity"
	TokenLookup string   `toml:"tokenLookup"` // token 查找位置，默认 "header:Authorization"
	SkipPaths   []string `toml:"skipPaths"`   // 跳过认证的路径列表
}

func DefaultConfig() Config {
	return Config{
		Realm:       "jwt",
		Timeout:     3600,
		MaxRefresh:  7200,
		IdentityKey: "identity",
		TokenLookup: "header:Authorization",
		SkipPaths:   []string{},
	}
}
