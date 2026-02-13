package middleware

import (
	"time"

	"github.com/CenJIl/base/logger"
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 创建日志中间件
//
// 使用 logger.GetLogger() 记录请求和响应
// Debug 模式：记录请求头、响应头和完整响应体
// 非 Debug 模式：只记录请求头和响应头
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()
		method := c.Request.Method

		logger.Debugf("[Request] %s %s?%s from %s", method, path, query, clientIP)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		length := c.Writer.Size()

		logger.Debugf("[Response] %s %s -> %d (Size: %d, Latency: %v)",
			method, path, status, length, latency)
	}
}
