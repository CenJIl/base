package main

import (
	"fmt"

	"github.com/CenJIl/base/cfg"
	"github.com/CenJIl/base/logger"
	"github.com/CenJIl/base/web"
)

type AppConfig struct {
	AppName string `toml:"appName"`
	Port    int    `toml:"port"`
	Debug   bool   `toml:"debug"`

	Web web.WebBaseConfig `toml:"web"`
}

func main() {
	defaultConfig := []byte(`appName = "MyApp"
port = 8080
debug = true

[web]
localePath = "./locales"
defaultLang = "zh-CN"
logLevel = "info"`)

	cfg.InitConfig[AppConfig](defaultConfig)
	config := cfg.GetCfg[AppConfig]()

	logger.Infof("服务启动: %s", config.AppName)

	engine := web.NewGin(config.Web)

	addr := fmt.Sprintf(":%d", config.Port)
	logger.Infof("服务监听在 %s", addr)
	if err := engine.Run(addr); err != nil {
		logger.Errorf("服务启动失败: %v", err)
	}
}
