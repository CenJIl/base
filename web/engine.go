package web

import (
	"context"
	"fmt"
	"time"

	"github.com/CenJIl/base/cfg"
	"github.com/CenJIl/base/logger"
	"github.com/CenJIl/base/web/cache"
	"github.com/CenJIl/base/web/database"
	"github.com/CenJIl/base/web/middleware"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	corsMiddleware "github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/i18n"
	jwtMiddleware "github.com/hertz-contrib/jwt"
	swaggerMiddleware "github.com/hertz-contrib/swagger"
)

// NewServer 创建 Hertz 服务器
//
// 配置完全由用户通过 config.toml 控制，此函数只负责：
// 1. 读取用户配置
// 2. 初始化已启用的子模块（DB/Redis/i18n）
// 3. 集成官方中间件（CORS/JWT/Swagger/i18n）
// 4. 返回可用的服务器实例
//
// # Generic parameter T 是用户的配置结构体类型，必须内嵌 web.Config
//
// 注意：
//   - 必须先调用 cfg.LoadConfig[AppConfig]() 或 cfg.InitConfig[AppConfig](defaultConfig)
//   - 配置文件中为零值的字段会被跳过（如 Port = 0 会导致 panic）
//
// Example:
//
//	type AppConfig struct {
//	    AppName string `toml:"appName"`
//	    web.Config  // 必须内嵌
//	}
//
//	cfg.LoadConfigFromDefaultPath[AppConfig]()
//	h := web.NewServer[AppConfig]()
func NewServer[T any]() *server.Hertz {
	// Load config from cfg package
	userCfg := cfg.GetCfg[T]()
	if userCfg == nil {
		panic("配置未初始化，请先调用 cfg.LoadConfig[AppConfig]() 或 cfg.InitConfig[AppConfig](defaultConfig)")
	}

	// Extract web config from embedded Config field
	webCfg := extractWebConfig(*userCfg)

	// 验证必要配置
	if webCfg.Port == 0 {
		panic("配置错误: web.port 不能为 0 或空，请在 config.toml 中设置 [web] port = 8080")
	}

	// Apply log level
	if webCfg.LogLevel != "" {
		logger.UpdateLogLevel(webCfg.LogLevel)
	}

	// Initialize i18n (使用官方 hertz-contrib/i18n)
	if webCfg.LocalePath != "" {
		if err := i18n.LoadI18n(webCfg.LocalePath, webCfg.DefaultLang); err != nil {
			logger.Warnf("[I18n] 加载失败: %v", err)
		} else {
			logger.Infof("[I18n] 已加载: %s (默认: %s)", webCfg.LocalePath, webCfg.DefaultLang)
		}
	}

	// Initialize database (如果配置了 driver)
	if webCfg.Database.Driver != "" {
		if err := database.InitDB(webCfg.Database); err != nil {
			panic(fmt.Errorf("数据库初始化失败: %w", err))
		}
		logger.Infof("[DB] 已连接: %s@%s:%d/%s",
			webCfg.Database.User, webCfg.Database.Host,
			webCfg.Database.Port, webCfg.Database.DBName)
	} else {
		logger.Info("[DB] 未配置 (database.driver 为空)")
	}

	// Initialize Redis (如果配置了 address)
	if webCfg.Redis.Address != "" {
		if err := cache.InitRedis(webCfg.Redis); err != nil {
			panic(fmt.Errorf("Redis 初始化失败: %w", err))
		}
		logger.Infof("[Redis] 已连接: %s", webCfg.Redis.Address)
	} else {
		logger.Info("[Redis] 未配置 (redis.address 为空)")
	}

	// Create Hertz server
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf(":%d", webCfg.Port)),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(15*time.Second),
		server.WithIdleTimeout(60*time.Second),
	)

	// ========== 注册全局中间件（按顺序） ==========

	// 1. 请求 ID 中间件（最外层，先生成）
	h.Use(middleware.RequestIDMiddleware())

	// 2. 安全头中间件
	h.Use(middleware.SecurityHeadersMiddleware())

	// 3. 全局异常处理
	h.Use(ExceptionHandler())

	// 4. 官方 i18n 中间件
	if webCfg.LocalePath != "" {
		h.Use(i18n.Locale())
	}

	// 5. 官方 CORS 中间件
	h.Use(corsMiddleware.CORS(
		// 允许所有源（生产环境应该限制）
		corsMiddleware.WithAllowOrigins([]string{"*"}),
		corsMiddleware.WithAllowMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		corsMiddleware.WithAllowHeaders([]string{"Content-Type", "Authorization"}),
		corsMiddleware.WithCredentials(true), // 允许携带 Cookie
	))

	// 6. 官方 JWT 中间件（后续需要配置 skipPaths）
	// h.Use(jwtMiddleware.HertzJWTMiddleware(...))

	// 7. 官方 Swagger 中间件（开发环境启用）
	// h.Use(swaggerMiddleware.Swagger(...))

	// Register static file serving (如果配置了 upload 路径和 URL 前缀）
	if webCfg.Upload.UploadPath != "" && webCfg.Upload.URLPrefix != "" {
		h.Static(webCfg.Upload.URLPrefix, webCfg.Upload.UploadPath)
		logger.Infof("[Static] %s -> %s", webCfg.Upload.URLPrefix, webCfg.Upload.UploadPath)
	}

	// Health check endpoint
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, utils.H{
			"code":    0,
			"message": "success",
			"data":    nil,
		})
	})

	return h
}

// MustRun 启动服务器（阻塞直到收到信号）
//
// # Generic parameter T 是用户的配置结构体类型
//
// Example:
//
//	h := web.NewServer[AppConfig]()
//	web.MustRun[AppConfig](h)
func MustRun[T any](h *server.Hertz) {
	userCfg := cfg.GetCfg[T]()
	if userCfg == nil {
		panic("配置未初始化，请先调用 cfg.LoadConfig[AppConfig]() 或 cfg.InitConfig[AppConfig](defaultConfig)")
	}

	webCfg := extractWebConfig(*userCfg)
	addr := fmt.Sprintf(":%d", webCfg.Port)

	logger.Infof("[HTTP] 服务监听: %s", addr)
	if err := h.Run(); err != nil {
		logger.Errorf("[HTTP] 启动失败: %v", err)
		panic(err)
	}
}

// GetPort 获取配置的端口号
//
// # Generic parameter T 是用户的配置结构体类型
//
// Example:
//
//	port := web.GetPort[AppConfig]()
func GetPort[T any]() int {
	userCfg := cfg.GetCfg[T]()
	if userCfg == nil {
		return 0 // 未初始化返回 0
	}
	webCfg := extractWebConfig(*userCfg)
	return webCfg.Port
}

// extractWebConfig 从用户配置中提取内嵌的 web.Config
func extractWebConfig(userCfg any) web.Config {
	type WebConfigEmbedder interface {
		GetWebConfig() web.Config
	}

	if embedder, ok := userCfg.(WebConfigEmbedder); ok {
		return embedder.GetWebConfig()
	}

	// 如果没有实现接口，返回零值
	return web.Config{}
}
