package common

import "log"

// Logger 通用日志接口
//
// 定义了日志记录器的基本行为，支持依赖注入和解耦
// 实现此接口的结构体可以作为 cfg 等模块的日志记录器
type Logger interface {
	Infof(format string, v ...any)  // 格式化输出 INFO 级别日志
	Errorf(format string, v ...any) // 格式化输出 ERROR 级别日志
}

// DefaultLog 默认日志实现
//
// 使用标准库 log 包实现的简单日志记录器
// INFO 级别日志添加 "INFO  " 前缀
// ERROR 级别日志添加 "ERROR " 前缀
type DefaultLog struct{}

// Infof 格式化输出 INFO 级别日志
//
// 使用标准库 log.Printf 输出日志，自动添加 "INFO  " 前缀
func (l *DefaultLog) Infof(format string, v ...any) {
	log.Printf("INFO  "+format, v...)
}

// Errorf 格式化输出 ERROR 级别日志
//
// 使用标准库 log.Printf 输出日志，自动添加 "ERROR " 前缀
func (l *DefaultLog) Errorf(format string, v ...any) {
	log.Printf("ERROR "+format, v...)
}
