package web

import (
	"github.com/CenJIl/base/web/cache"
	"github.com/CenJIl/base/web/database"
)

// DatabaseConfig 数据库配置（类型别名）
type DatabaseConfig = database.DatabaseConfig

// RedisConfig Redis 配置（类型别名）
type RedisConfig = cache.RedisConfig

// Config Web 基础配置
//
// 使用者必须在配置结构体中内嵌此配置
// 例如：
//
//	type AppConfig struct {
//	    AppName string `toml:"appName"`
//	    Port    int    `toml:"port"`
//	    web.Config  // 必须内嵌
// }

type Config struct {
	LocalePath  string `toml:"localePath"`  // 本地化文件路径，默认 "./locales"
	DefaultLang string `toml:"defaultLang"` // 默认语言，默认 "zh-CN"
	LogLevel    string `toml:"logLevel"`    // 日志级别，默认 "info"
	Port        int    `toml:"port"`        // HTTP 监听端口
	Upload      UploadConfig `toml:"upload"`     // 文件上传配置
	Database    DatabaseConfig `toml:"database"`  // 数据库配置（可选）
	Redis      RedisConfig    `toml:"redis"`      // Redis 配置（可选）
}

// extractWebConfig 从用户配置中提取内嵌的 web.Config
// 使用类型断言获取内嵌字段
func extractWebConfig(userCfg any) Config {
	// 尝试直接断言为 Config（当用户直接内嵌 Config 时）
	if c, ok := userCfg.(Config); ok {
		return c
	}

	// 返回默认配置（当无法提取时使用默认值）
	return Config{
		LocalePath: "./locales",
		DefaultLang: "zh-CN",
		LogLevel:   "info",
		Port:       8080,
		Upload: UploadConfig{
			MaxFileSize: 10 * 1024 * 1024, // 10MB
			AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf"},
			UploadPath:  "./uploads",
			URLPrefix:   "/uploads",
		},
		Database: DatabaseConfig{
			Driver: "", // 默认不启用数据库
			Host:   "localhost",
			Port:   3306,
			User:   "root",
			Password: "",
			DBName: "myapp",
			MaxOpen: 100,
			MaxIdle: 10,
		},
		Redis: RedisConfig{
			Address: "", // 默认不启用 Redis
			Password: "",
			DB:      0,
		},
	}
}
