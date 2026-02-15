package web

import (
	"reflect"

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
//
// Example:
//
//	type AppConfig struct {
//	    AppName string `toml:"appName"`
//	    Port    int    `toml:"port"`
//	    web.Config  // 必须内嵌
//	}
type Config struct {
	LocalePath  string         `toml:"localePath"`  // 本地化文件路径
	DefaultLang string         `toml:"defaultLang"` // 默认语言
	LogLevel    string         `toml:"logLevel"`    // 日志级别
	Port        int            `toml:"port"`        // HTTP 监听端口
	Upload      UploadConfig   `toml:"upload"`      // 文件上传配置
	Database    DatabaseConfig `toml:"database"`    // 数据库配置（可选）
	Redis       RedisConfig    `toml:"redis"`       // Redis 配置（可选）
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxFileSize int64    `toml:"maxFileSize"` // 单文件最大大小（字节）
	AllowedExts []string `toml:"allowedExts"` // 允许的扩展名
	UploadPath  string   `toml:"uploadPath"`  // 上传保存路径
	URLPrefix   string   `toml:"urlPrefix"`   // 访问 URL 前缀
}

// extractWebConfig 从用户配置中提取内嵌的 web.Config
//
// 使用反射提取内嵌字段
//
// 如果用户配置中未内嵌 Config，返回零值
func extractWebConfig(userCfg any) Config {
	// 获取值的反射
	val := reflect.ValueOf(userCfg)

	// 如果是指针，解引用
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 遍历所有字段，查找内嵌的 Config
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 查找匿名（内嵌）字段且类型为 Config
		if field.Anonymous && field.Type == reflect.TypeOf(Config{}) {
			// 找到内嵌的 Config，直接返回
			if fieldValue.CanInterface() {
				if cfg, ok := fieldValue.Interface().(Config); ok {
					return cfg
				}
			}
		}
	}

	// 没找到内嵌 Config，返回零值
	return Config{}
}
