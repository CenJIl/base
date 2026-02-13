package web

import (
	"github.com/CenJIl/base/web/middleware"
	"github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
)

// ExampleHandler 示例处理器
//
// 展示如何使用 i18n 和统一响应格式
func ExampleHandler(c *gin.Context) {
	name := c.Query("name")

	c.JSON(200, middleware.Success(map[string]string{
		"greeting": i18n.MustGetMessage(c, "welcomeWithName"),
		"name":     name,
	}))
}
