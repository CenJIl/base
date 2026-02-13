package middleware

import (
	"github.com/CenJIl/base/logger"
	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 创建恢复中间件
//
// 捕获 panic 并转换为统一的错误响应
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[PANIC] %v", r)
				c.JSON(500, Error(500, "Internal Server Error", nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}
