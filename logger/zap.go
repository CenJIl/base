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

// init 包初始化时自动执行，完整初始化 logger
func init() {
	atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapSugarLogger = createFullLogger()
}

// createFullLogger 创建完整的 logger（控制台 + 可选文件）
func createFullLogger() *zap.SugaredLogger {
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
		// 控制台输出（彩色）
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

func GetLogger() *zap.SugaredLogger {
	return zapSugarLogger
}

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
