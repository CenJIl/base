package web

import (
	"context"
	"time"

	"github.com/CenJIl/base/logger"
	"github.com/CenJIl/base/web/middleware"
	"github.com/cloudwego/hertz/pkg/app"
)

// RecoveryMiddleware 恢复中间件
//
// 捕获 panic 并转换为统一的错误响应
func RecoveryMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[PANIC] %v", r)
				result := Fail(500, "Internal server error")
				result.TraceID = middleware.GetRequestID(c)
				c.JSON(500, result)
				c.Abort()
			}
		}()
		c.Next(ctx)
	}
}

// LoggerMiddleware 日志中间件
//
// 记录每个请求的详细信息
func LoggerMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := c.Path()
		method := string(c.Method())
		clientIP := c.ClientIP()

		logger.Debugf("[Request] %s %s from %s", method, path, clientIP)

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()
		logger.Debugf("[Response] %s %s -> %d (Latency: %v)",
			method, path, status, latency)
	}
}

// getHTTPStatus 根据业务码获取 HTTP 状态码
func getHTTPStatus(code int) int {
	switch code / 100 {
	case 1, 2, 3, 4:
		// 1xx-4xx 错误直接返回对应状态码
		if code >= 400 && code < 600 {
			return code
		}
		return 400
	case 5:
		return 500
	default:
		return 500
	}
}

// ExceptionHandler 全局异常处理器（类似 Spring Boot 的 @RestControllerAdvice）
func ExceptionHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				result := Result{}
				switch err := r.(type) {
				case *HTTPException:
					// HTTP 异常
					result = Fail(err.Code, err.Message)
					result.TraceID = middleware.GetRequestID(c)
					c.JSON(err.HTTPStatus, result)
					c.Abort()
					return

				case *Exception:
					// 业务异常
					result = Fail(err.Code, err.Message)
					result.TraceID = middleware.GetRequestID(c)
					c.JSON(getHTTPStatus(err.Code), result)
					c.Abort()
					return

				default:
					logger.Errorf("[PANIC] Unhandled error: %v", err)
					result = Fail(500, "Internal server error")
					result.TraceID = middleware.GetRequestID(c)
					c.JSON(500, result)
					c.Abort()
				}
			}
		}()
		c.Next(ctx)
	}
}
