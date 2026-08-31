package web

import (
	"context"
	"errors"
	"reflect"

	"github.com/CenJIl/base/cfg"
	"github.com/CenJIl/base/web/cache"
	"github.com/CenJIl/base/web/database"
)

type DatabaseConfig = database.DatabaseConfig
type RedisConfig = cache.RedisConfig

// Config contains the optional modules and defaults used by NewServer.
type Config struct {
	LocalePath  string         `toml:"localePath"`
	DefaultLang string         `toml:"defaultLang"`
	LogLevel    string         `toml:"logLevel"`
	Port        int            `toml:"port"`
	Upload      UploadConfig   `toml:"upload"`
	Database    DatabaseConfig `toml:"database"`
	Redis       RedisConfig    `toml:"redis"`
	CORS        CORSConfig     `toml:"cors"`
	Security    SecurityConfig `toml:"security"`
}

type CORSConfig struct {
	AllowOrigins     []string `toml:"allowOrigins"`
	AllowMethods     []string `toml:"allowMethods"`
	AllowHeaders     []string `toml:"allowHeaders"`
	AllowCredentials bool     `toml:"allowCredentials"`
}

type SecurityConfig struct {
	AllowedOrigins []string `toml:"allowedOrigins"`
	MaxBodySize    int64    `toml:"maxBodySize"`
}

type UploadConfig struct {
	MaxFileSize int64    `toml:"maxFileSize"`
	AllowedExts []string `toml:"allowedExts"`
	UploadPath  string   `toml:"uploadPath"`
	URLPrefix   string   `toml:"urlPrefix"`
}

func (c Config) DefaultCORS() CORSConfig {
	if len(c.CORS.AllowOrigins) == 0 {
		c.CORS.AllowOrigins = []string{"http://localhost:3000"}
	}
	if c.CORS.AllowCredentials {
		for _, origin := range c.CORS.AllowOrigins {
			if origin == "*" {
				c.CORS.AllowCredentials = false
				break
			}
		}
	}
	if len(c.CORS.AllowMethods) == 0 {
		c.CORS.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(c.CORS.AllowHeaders) == 0 {
		c.CORS.AllowHeaders = []string{"Content-Type", "Authorization", "X-Request-ID"}
	}
	return c.CORS
}
func (c Config) DefaultSecurity() SecurityConfig {
	if c.Security.MaxBodySize <= 0 {
		c.Security.MaxBodySize = 10 << 20
	}
	return c.Security
}

func extractWebConfig(userCfg any) Config {
	val := reflect.ValueOf(userCfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if !val.IsValid() || val.Kind() != reflect.Struct {
		return Config{}
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous && field.Type == reflect.TypeOf(Config{}) {
			if value := val.Field(i); value.CanInterface() {
				if cfg, ok := value.Interface().(Config); ok {
					return cfg
				}
			}
		}
	}
	return Config{}
}

// Shutdown closes the Hertz server and every process-wide optional resource.
func Shutdown(ctx context.Context, h interface{ Shutdown(context.Context) error }) error {
	var serverErr error
	if h != nil {
		serverErr = h.Shutdown(ctx)
	}
	CloseRateLimiter()
	return errors.Join(serverErr, database.Close(), cache.Close(), cfg.Close())
}
