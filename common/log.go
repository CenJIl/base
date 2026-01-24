package common

import "log"

// Logger 通用日志接口
type Logger interface {
	Infof(format string, v ...any)
	Errorf(format string, v ...any)
}

// DefaultLog 默认日志实现，使用标准库 log
type DefaultLog struct{}

func (l *DefaultLog) Infof(format string, v ...any) {
	log.Printf("INFO  "+format, v...)
}

func (l *DefaultLog) Errorf(format string, v ...any) {
	log.Printf("ERROR "+format, v...)
}
