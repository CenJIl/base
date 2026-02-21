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
	hertzI18n "github.com/hertz-contrib/i18n"
	_ "github.com/hertz-contrib/jwt"
	_ "github.com/hertz-contrib/swagger"
)

// NewServer 创建 Hertz 服务器
//
// 配置完全由用户通过配置文件控制，此函数负责：
// 1. 读取用户配置（从 app.toml 或指定路径）
// 2. 初始化已启用的子模块（DB/Redis/i18n）
// 3. 集成官方中间件（CORS/JWT/Swagger/i18n）
// 4. 返回可用的服务器实例
//
// # Generic parameter T 是用户的配置结构体类型，必须内嵌 web.Config
//
// # Parameters
//
//	configPath - 配置文件路径，可选。默认为 "app.toml"
//
// 注意：
//   - 配置文件中为零值的字段会被跳过（如 Port = 0 会导致 panic）
//
// Example:
//
//	type AppConfig struct {
//	    AppName string `toml:"appName"`
//	    web.Config  // 必须内嵌
//	}
//
//	// 无参数 - 默认读取 ./app.toml
//	h := web.NewServer[AppConfig]()
//
//	// 有参数 - 读取自定义路径
//	h := web.NewServer[AppConfig]("config/app.toml")
func NewServer[T any](configPath ...string) *server.Hertz {
	// 确定配置文件路径
	configFile := "app.toml"
	if len(configPath) > 0 && configPath[0] != "" {
		configFile = configPath[0]
	}

	// 加载配置
	if err := cfg.LoadConfig[T](configFile); err != nil {
		panic(fmt.Errorf("配置加载失败: %w", err))
	}

	// Load config from cfg package
	userCfg := cfg.GetCfg[T]()
	if userCfg == nil {
		panic("配置未初始化，请先调用 cfg.LoadConfig 或使用 NewServer 的自动加载功能")
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
		h.Use(hertzI18n.Localize())
	}

	// 5. 官方 CORS 中间件
	h.Use(corsMiddleware.New(corsMiddleware.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

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
		panic("配置未初始化，请先调用 web.NewServer[AppConfig]()")
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
