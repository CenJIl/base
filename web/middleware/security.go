package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// SecurityHeadersMiddleware sets conservative security response headers.
func SecurityHeadersMiddleware() app.HandlerFunc {
	return SecurityHeadersMiddlewareWithCSP("default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; object-src 'none")
}

// SecurityHeadersMiddlewareWithCSP uses the supplied Content-Security-Policy.
func SecurityHeadersMiddlewareWithCSP(csp string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		if csp != "" {
			c.Header("Content-Security-Policy", csp)
		}
		c.Next(ctx)
	}
}
