package web

import (
	"github.com/CenJIl/base/logger"
	"github.com/CenJIl/base/web/middleware"
	"github.com/gin-gonic/gin"
)

func NewGin(webCfg WebBaseConfig) *gin.Engine {
	r := gin.New()

	logger.UpdateLogLevel(webCfg.LogLevel)

	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.I18nMiddleware(webCfg.LocalePath, webCfg.DefaultLang))
	r.Use(middleware.LoggerMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, middleware.Success(nil))
	})

	return r
}
