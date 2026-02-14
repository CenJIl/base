package web

import (
	"context"
	"net/http"

	"github.com/CenJIl/base/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

// WrapHandler 响应包装器（类似 AOP）
//
// 自动捕获 panic 和 error，转换为统一响应格式
// 支持直接返回 error 的 handler 函数
//
// 使用方式：
//   h.POST("/users", web.WrapHandler(createUser))
//
//   func createUser(ctx context.Context, c *app.RequestContext) error {
//       if err := validateUser(c); err != nil {
//           return web.BadRequest("Invalid user data")
//       }
//       return nil  // 自动返回 Success(nil)
//   }
func WrapHandler(h func(ctx context.Context, c *app.RequestContext) error) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 执行 handler 并捕获 panic
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[PANIC] Handler panic: %v", r)

				switch err := r.(type) {
				case *HTTPException:
					c.JSON(err.HTTPStatus, Fail(err.Code, err.Message))
					c.Abort()
				case *Exception:
					c.JSON(getHTTPStatus(err.Code), Fail(err.Code, err.Message))
					c.Abort()
				case error:
					c.JSON(http.StatusInternalServerError, Fail(500, err.Error()))
					c.Abort()
				default:
					c.JSON(http.StatusInternalServerError, Fail(500, "Internal server error"))
					c.Abort()
				}
			}
		}()

		// 调用 handler
		if err := h(ctx, c); err != nil {
			// 处理错误
			switch e := err.(type) {
			case *HTTPException:
				c.JSON(e.HTTPStatus, Fail(e.Code, e.Message))
				c.Abort()
			case *Exception:
				c.JSON(getHTTPStatus(e.Code), Fail(e.Code, e.Message))
				c.Abort()
			default:
				logger.Errorf("[ERROR] Handler error: %v", err)
				c.JSON(http.StatusInternalServerError, Fail(500, err.Error()))
				c.Abort()
			}
			return
		}

		// 如果没有写入响应且没有错误，自动返回 Success
		if c.Response.StatusCode() == 0 {
			c.JSON(200, Success(nil))
		}
	}
}
