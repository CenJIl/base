package middleware

import (
	"os"
	"path/filepath"

	"github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

func I18nMiddleware(localePath, defaultLang string) gin.HandlerFunc {
	exePath, err := os.Executable()
	if err != nil {
		panic("获取可执行文件路径失败: " + err.Error())
	}
	exeDir := filepath.Dir(exePath)
	localeDir := filepath.Join(exeDir, localePath)

	defaultTag, err := language.Parse(defaultLang)
	if err != nil {
		panic("解析默认语言失败: " + err.Error())
	}

	// 使用正确的语言标签，确保与文件名匹配
	acceptLangs := []language.Tag{defaultTag}
	// 添加英文支持
	if enTag := language.Make("en-US"); enTag != defaultTag {
		acceptLangs = append(acceptLangs, enTag)
	}

	return i18n.Localize(
		i18n.WithBundle(&i18n.BundleCfg{
			RootPath:         localeDir,
			AcceptLanguage:   acceptLangs,
			DefaultLanguage:  defaultTag,
			UnmarshalFunc:    toml.Unmarshal,
			FormatBundleFile: "toml",
		}),
	)
}
