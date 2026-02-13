package logger

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	zapSugarLogger *zap.SugaredLogger
	atomicLevel    zap.AtomicLevel
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

func init() {
	atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	baseEncoderConfig := zapcore.EncoderConfig{
		TimeKey:       "t",
		LevelKey:      "l",
		NameKey:       "",
		CallerKey:     "",
		FunctionKey:   "",
		MessageKey:    "m",
		StacktraceKey: "",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("15:04:05.000"))
		},
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}

	consoleEncoderConfig := baseEncoderConfig
	consoleEncoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString(colorBlue + "DEBUG" + colorReset)
		case zapcore.InfoLevel:
			enc.AppendString(colorGreen + "INFO " + colorReset)
		case zapcore.WarnLevel:
			enc.AppendString(colorYellow + "WARN " + colorReset)
		case zapcore.ErrorLevel:
			enc.AppendString(colorRed + "ERROR" + colorReset)
		default:
			enc.AppendString(l.CapitalString())
		}
	}

	fileEncoderConfig := baseEncoderConfig
	fileEncoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.InfoLevel:
			enc.AppendString("INFO ")
		case zapcore.WarnLevel:
			enc.AppendString("WARN ")
		default:
			enc.AppendString(l.CapitalString())
		}
	}

	coreConfigs := []zapcore.Core{
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEncoderConfig), zapcore.AddSync(os.Stdout), atomicLevel),
	}

	if isWindowsService() {
		exePath, err := os.Executable()
		if err != nil {
			panic("获取可执行文件路径失败: " + err.Error())
		}
		logDir := filepath.Join(filepath.Dir(exePath), "logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			panic("创建日志目录失败: " + err.Error())
		}
		lumberjackLogger := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "app.log"),
			MaxSize:    20,
			MaxBackups: 10,
			MaxAge:     30,
			LocalTime:  true,
			Compress:   true,
		}
		coreConfigs = append(coreConfigs, zapcore.NewCore(zapcore.NewConsoleEncoder(fileEncoderConfig), zapcore.AddSync(lumberjackLogger), atomicLevel))
	}

	zapSugarLogger = zap.New(zapcore.NewTee(coreConfigs...)).Sugar()
}

// GetLogger 返回全局日志记录器实例
//
// 返回的是一个 zap.SugaredLogger 实例，支持结构化日志记录。
// 此日志记录器是全局唯一的，在包初始化时自动创建。
//
// 输出目标
//   - 控制台：始终输出到标准输出，带彩色级别标识
//   - 文件：当作为 Windows 服务运行时，自动输出到 logs/app.log
//
// 返回值
//
//	*zap.SugaredLogger - Zap 的结构化日志记录器，支持多种日志方法
//
// 注意事项
//   - 日志记录器在包导入时自动初始化，无需手动调用
//   - 返回的日志记录器是全局单例，所有调用共享同一个实例
//   - 文件日志仅在 Windows 服务模式下启用
//   - 日志文件会自动轮转（最大 20MB，保留 30 天，最多 10 个备份）
//
// 示例
//
//	logger := logger.GetLogger()
//	logger.Info("应用启动", "version", "1.0.0")
//	logger.Errorf("操作失败", "error", err)
func GetLogger() *zap.SugaredLogger {
	return zapSugarLogger
}

// UpdateLogLevel 动态更新日志级别
//
// 允许在运行时动态调整日志级别，无需重启程序。
// 支持的级别：debug, info, warn, error（不区分大小写）
//
// 参数
//
//	level - 目标日志级别字符串，支持 "debug", "info", "warn", "error"（大小写不敏感）
//
// 注意事项
//   - 级别字符串会自动 trim 空白和转换为小写
//   - 如果传入无效的级别，记录错误日志但不修改当前级别
//   - 修改成功后会记录日志，显示旧级别到新级别的变更
//   - 日志级别变更立即生效，影响后续所有日志输出
//
// 示例
//
//	logger.UpdateLogLevel("debug")  // 开启 debug 日志
//	logger.UpdateLogLevel("INFO")   // 切换到 info 级别
func UpdateLogLevel(level string) {
	var l zapcore.Level
	err := l.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(level))))
	if err != nil {
		zapSugarLogger.Errorf("无法解析日志级别: %s", level)
		return
	}
	if l != atomicLevel.Level() {
		oldLevel := atomicLevel.Level().String()
		atomicLevel.SetLevel(l)
		zapSugarLogger.Infof("日志级别已更新: %s -> %s", strings.ToUpper(oldLevel), strings.ToUpper(l.String()))
	}
}

func Debug(msg string) { zapSugarLogger.Debug(msg) }
func Info(msg string)  { zapSugarLogger.Info(msg) }
func Warn(msg string)  { zapSugarLogger.Warn(msg) }
func Error(msg string) { zapSugarLogger.Error(msg) }

func Debugf(format string, args ...any) { zapSugarLogger.Debugf(format, args...) }
func Infof(format string, args ...any)  { zapSugarLogger.Infof(format, args...) }
func Warnf(format string, args ...any)  { zapSugarLogger.Warnf(format, args...) }
func Errorf(format string, args ...any) { zapSugarLogger.Errorf(format, args...) }
