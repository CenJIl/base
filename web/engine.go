package web

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/CenJIl/base/cfg"
	"github.com/CenJIl/base/logger"
	"github.com/CenJIl/base/web/cache"
	"github.com/CenJIl/base/web/database"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ensureDir 确保目录存在
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// NewServer creates a configured Hertz server
//
// Automatically loads config from cfg package and applies web.Config
// Auto configures: log level, recovery, logger, CORS middleware
//
// Generic parameter T is user's config struct type, must embed web.Config
//
// Returns configured *hertz.Server, register routes and use Spin() to start
//
// Example:
//
//	type AppConfig struct {
//	    AppName string `toml:"appName"`
//	    web.Config  // Must embed web.Config
//	}
//
//	cfg.InitConfig[AppConfig](defaultConfig)
//	h := web.NewServer[AppConfig]()
//	h.GET("/api", myHandler)
//	h.Spin()
func NewServer[T any]() *server.Hertz {
	// Load config from cfg package
	userCfg := cfg.GetCfg[T]()
	if userCfg == nil {
		panic("config not initialized, call cfg.InitConfig[T]() first")
	}

	// Extract web config from embedded Config field
	webCfg := extractWebConfig(*userCfg)

	// Apply log level
	logger.UpdateLogLevel(webCfg.LogLevel)

	// Initialize database (if configured)
	if webCfg.Database.Driver != "" {
		if err := database.InitDB(webCfg.Database); err != nil {
			panic(fmt.Sprintf("Failed to init database: %v", err))
		}
		logger.Infof("Database connected: %s@%s:%d/%s",
			webCfg.Database.User, webCfg.Database.Host,
			webCfg.Database.Port, webCfg.Database.DBName)
	}

	// Initialize Redis (if configured)
	if webCfg.Redis.Address != "" {
		if err := cache.InitRedis(webCfg.Redis); err != nil {
			panic(fmt.Sprintf("Failed to init redis: %v", err))
		}
		logger.Infof("Redis connected: %s", webCfg.Redis.Address)
	}

	// Create Hertz server
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf(":%d", webCfg.Port)),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(15*time.Second),
		server.WithIdleTimeout(60*time.Second),
	)

	// Register global exception handler
	h.Use(ExceptionHandler())

	// Register static file serving (if upload path is configured)
	if webCfg.Upload.UploadPath != "" && webCfg.Upload.URLPrefix != "" {
		// 确保上传目录存在
		if err := ensureDir(webCfg.Upload.UploadPath); err != nil {
			logger.Warnf("Failed to create upload directory: %v", err)
		} else {
			h.Static(webCfg.Upload.URLPrefix, webCfg.Upload.UploadPath)
			logger.Infof("Static file serving: %s -> %s", webCfg.Upload.URLPrefix, webCfg.Upload.UploadPath)
		}
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

// MustRun starts Hertz server (panics on error)
//
// Blocks until shutdown signal received
// Auto handles SIGINT and SIGTERM signals
//
// Generic parameter T is user's config struct type
//
// Example:
//
//	h := web.NewServer[AppConfig]()
//	web.MustRun[AppConfig](h)
func MustRun[T any](h *server.Hertz) {
	userCfg := cfg.GetCfg[T]()
	if userCfg == nil {
		panic("config not initialized, call cfg.InitConfig[T]() first")
	}

	webCfg := extractWebConfig(*userCfg)
	addr := fmt.Sprintf(":%d", webCfg.Port)

	logger.Infof("HTTP server listening on %s", addr)
	if err := h.Run(); err != nil {
		logger.Errorf("HTTP server failed to start: %v", err)
		panic(err)
	}
}

// GetPort gets configured port number
//
// Generic parameter T is user's config struct type
func GetPort[T any]() int {
	userCfg := cfg.GetCfg[T]()
	if userCfg == nil {
		return 8080 // Default port
	}
	webCfg := extractWebConfig(*userCfg)
	return webCfg.Port
}
