package main

import (
	"context"

	"github.com/CenJIl/base/web"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type AppConfig struct {
	AppName string `toml:"appName"`
	web.Config
}

func main() {
	h := web.NewServer[AppConfig]()

	h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, web.Success(map[string]string{
			"message": "Hello, World!",
		}))
	})

	web.MustRun[AppConfig](h)
}
