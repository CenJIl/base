package web_test

import (
	"context"

	"github.com/CenJIl/base/web"
	"github.com/cloudwego/hertz/pkg/app"
)

// ExampleI18nUsage i18n 使用示例
//
// 展示如何使用高性能的 i18n 功能
func ExampleI18nUsage() {
	// 1. 初始化 i18n（在 main.go 或启动时）
	// 预加载所有翻译到内存（性能优化）
	web.InitI18n(map[string]map[string]map[string]string{
		"zh-CN": {
			"welcome":       "欢迎",
			"user.not_found": "用户不存在",
			"login.success": "登录成功",
			"hello.user":   "你好 %s",
		},
		"en-US": {
			"welcome":       "Welcome",
			"user.not_found": "User not found",
			"login.success": "Login successful",
			"hello.user":   "Hello %s",
		},
	})

	// 2. 在 handler 中使用
	_ = func(ctx context.Context, c *app.RequestContext) {
		// 自动检测语言（从 ?lang= 或 Accept-Language）
		// 使用 web.T() 翻译
		msg := web.T(c, "welcome")
		_ = msg
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 翻译带参数的消息
		greeting := web.T(c, "hello.user", "张三")
		_ = greeting // "你好 张三"
	}

	_ = func(ctx context.Context, c *app.RequestContext) {
		// 翻译错误消息
		errMsg := web.T(c, "user.not_found")
		_ = errMsg // "用户不存在"
	}
}

// ExampleI18nRealWorld 实际项目使用
func ExampleI18nRealWorld() {
	// 在 main.go 或 web.NewServer 中初始化
	// web.InitI18n(translations)
	// web.I18nMiddleware("zh-CN")
}

// ExampleI18nHandler 在 handler 中使用
func ExampleI18nHandler() {
	_ = func(ctx context.Context, c *app.RequestContext) {
		// 直接使用 T() 翻译
		c.JSON(200, map[string]any{
			"message": web.T(ctx, "login.success"),
		})
	}
}
