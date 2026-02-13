package web

// WebBaseConfig Web 基础配置
//
// 使用者必须在配置结构体中内嵌此配置
// 例如：
//
//	type AppConfig struct {
//	    AppName string `toml:"appName"`
//	    Port    int    `toml:"port"`
//	    web.WebBaseConfig  // 必须内嵌
//	}
type WebBaseConfig struct {
	LocalePath  string `toml:"localePath"`  // 本地化文件路径，默认 "./locales"
	DefaultLang string `toml:"defaultLang"` // 默认语言，默认 "zh-CN"
	LogLevel    string `toml:"logLevel"`    // 日志级别，默认 "info"
}
