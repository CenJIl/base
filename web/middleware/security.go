package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// SecurityHeadersMiddleware 安全响应头中间件
//
// 设置安全相关的 HTTP 响应头，防止常见 Web 攻击
//
// 设置的头：
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - X-XSS-Protection: 1; mode=block
//   - Content-Security-Policy: default-src 'self'
//
// Example:
//
//	h.Use(middleware.SecurityHeadersMiddleware())
func SecurityHeadersMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 防止 MIME 类型嗅探（XSS 攻击）
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持（防止嵌入到 iframe）
		c.Header("X-Frame-Options", "DENY")

		// 启用 XSS 保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 内容安全策略（只允许同源资源）
		c.Header("Content-Security-Policy", "default-src 'self' http://https: 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; object-src 'none'")

		// 防止缓存敏感信息
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")

		c.Next(ctx)
	}
}
