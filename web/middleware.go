package web

import (
	"context"

	"github.com/CenJIl/base/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

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
				switch err := r.(type) {
				case *HTTPException:
					// HTTP 异常
					c.JSON(err.HTTPStatus, Fail(err.Code, err.Message))
					c.Abort()
					return

				case *Exception:
					// 业务异常
					c.JSON(getHTTPStatus(err.Code), Fail(err.Code, err.Message))
					c.Abort()
					return

				default:
					logger.Errorf("[PANIC] Unhandled error: %v", err)
					c.JSON(500, Fail(500, "Internal server error"))
					c.Abort()
				}
			}
		}()
		c.Next(ctx)
	}
}
