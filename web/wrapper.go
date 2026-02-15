package web

import (
	"context"
	"net/http"

	"github.com/CenJIl/base/logger"
	"github.com/CenJIl/base/web/middleware"
	"github.com/cloudwego/hertz/pkg/app"
)

// WrapHandler 响应包装器（类似 AOP）
//
// 自动捕获 panic 和 error，转换为统一响应格式
// 支持直接返回 error 的 handler 函数
// 自动为响应添加 TraceID
//
// 使用方式：
//
//	h.POST("/users", web.WrapHandler(createUser))
//
//	func createUser(ctx context.Context, c *app.RequestContext) error {
//	    if err := validateUser(c); err != nil {
//	        return web.BadRequest("Invalid user data")
//	    }
//	    return nil  // 自动返回 Success(nil)
//	}
func WrapHandler(h func(ctx context.Context, c *app.RequestContext) error) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 执行 handler 并捕获 panic
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[PANIC] Handler panic: %v", r)

				result := Result{}
				switch err := r.(type) {
				case *HTTPException:
					result = Fail(err.Code, err.Message)
					result.TraceID = middleware.GetRequestID(c)
					c.JSON(err.HTTPStatus, result)
					c.Abort()
				case *Exception:
					result = Fail(err.Code, err.Message)
					result.TraceID = middleware.GetRequestID(c)
					c.JSON(getHTTPStatus(err.Code), result)
					c.Abort()
				case error:
					result = Fail(500, err.Error())
					result.TraceID = middleware.GetRequestID(c)
					c.JSON(http.StatusInternalServerError, result)
					c.Abort()
				default:
					result = Fail(500, "Internal server error")
					result.TraceID = middleware.GetRequestID(c)
					c.JSON(http.StatusInternalServerError, result)
					c.Abort()
				}
			}
		}()

		// 调用 handler
		if err := h(ctx, c); err != nil {
			// 处理错误
			result := Result{}
			switch e := err.(type) {
			case *HTTPException:
				result = Fail(e.Code, e.Message)
				result.TraceID = middleware.GetRequestID(c)
				c.JSON(e.HTTPStatus, result)
				c.Abort()
			case *Exception:
				result = Fail(e.Code, e.Message)
				result.TraceID = middleware.GetRequestID(c)
				c.JSON(getHTTPStatus(e.Code), result)
				c.Abort()
			default:
				logger.Errorf("[ERROR] Handler error: %v", err)
				result = Fail(500, err.Error())
				result.TraceID = middleware.GetRequestID(c)
				c.JSON(http.StatusInternalServerError, result)
				c.Abort()
			}
			return
		}

		// 如果没有写入响应且没有错误，自动返回 Success
		if c.Response.StatusCode() == 0 {
			result := Success(nil)
			result.TraceID = middleware.GetRequestID(c)
			c.JSON(200, result)
		}
	}
}
