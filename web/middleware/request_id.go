package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

// RequestIDMiddleware 请求 ID 中间件
//
// 为每个请求生成唯一的请求 ID，方便日志追踪和问题排查
//
// 功能：
//   - 生成 UUID 作为请求 ID
//   - 设置到响应头 X-Request-ID
//   - 存储到上下文供后续使用
//
// Example:
//
//	h.Use(middleware.RequestIDMiddleware())
func RequestIDMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 生成请求 ID
		requestID := uuid.New().String()

		// 设置到响应头（前端可以拿到）
		c.Header("X-Request-ID", requestID)

		// 存储到上下文（Handler 可以使用）
		c.Set("request_id", requestID)

		c.Next(ctx)
	}
}

// GetRequestID 从上下文获取请求 ID
//
// Example:
//
//	requestID := middleware.GetRequestID(c)
func GetRequestID(c *app.RequestContext) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// RequestIDKey 是 context.Context 中存储请求 ID 的键
type RequestIDKey struct{}

// GetRequestIDFromContext 从 context.Context 获取请求 ID
// 用于 Success/Fail 等响应函数中无法直接访问 RequestContext 的场景
//
// Example:
//
//	requestID := middleware.GetRequestIDFromContext(ctx)
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}
