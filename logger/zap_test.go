package logger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	assert.NotNil(t, GetLogger())

	logger2 := GetLogger()
	assert.NotNil(t, logger2)
}

func TestUpdateLogLevel_ValidLevels(t *testing.T) {
	testCases := []string{
		"debug",
		"info",
		"warn",
		"error",
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
	}

	for _, level := range testCases {
		t.Run(level, func(t *testing.T) {
			assert.NotPanics(t, func() {
				UpdateLogLevel(level)
			})
		})
	}
}

func TestUpdateLogLevel_InvalidLevel(t *testing.T) {
	assert.NotPanics(t, func() {
		UpdateLogLevel("invalid")
	})

	assert.NotPanics(t, func() {
		UpdateLogLevel("")
	})
}

func TestUpdateLogLevel_TrimWhitespace(t *testing.T) {
	testCases := []string{
		" debug ",
		"info  ",
		"  warn",
		"\nerror\n",
	}

	for _, level := range testCases {
		t.Run(level, func(t *testing.T) {
			assert.NotPanics(t, func() {
				UpdateLogLevel(level)
			})
		})
	}
}

func TestUpdateLogLevel_CaseInsensitive(t *testing.T) {
	testCases := []string{
		"DEBUG",
		"debug",
		"DeBuG",
		"INFO",
		"info",
		"InFo",
		"WARN",
		"warn",
		"Warn",
		"ERROR",
		"error",
		"ErRoR",
	}

	for _, level := range testCases {
		t.Run(level, func(t *testing.T) {
			assert.NotPanics(t, func() {
				UpdateLogLevel(level)
			})
		})
	}
}

func TestLoggingMethods(t *testing.T) {
	assert.NotNil(t, GetLogger())

	assert.NotPanics(t, func() {
		Debug("Debug message")
		Info("Info message")
		Warn("Warn message")
		Error("Error message")
	})
}

func TestLoggingMethodsFormatted(t *testing.T) {
	assert.NotNil(t, GetLogger())

	assert.NotPanics(t, func() {
		Debugf("Debug %s", "formatted")
		Infof("Info %s", "formatted")
		Warnf("Warn %s", "formatted")
		Errorf("Error %s", "formatted")
	})
}

func TestLoggingMultipleArgs(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Infof("Multiple %s %d %v", "args", 42, true)
	})
}

func TestLoggingNoArgs(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Infof("No args")
	})
}

func TestLoggingWithEmptyMessage(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Info("")
		Debug("")
		Warn("")
		Error("")
	})
}

func TestLoggingSpecialCharacters(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Info("Message with % special % characters")
		Infof("Message with \n newline")
		Infof("Message with \t tab")
	})
}

func TestLoggingChineseCharacters(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Info("中文日志消息")
		Infof("格式化中文: %s", "测试")
		Warn("警告消息")
		Error("错误消息")
	})
}

func TestLoggingUnicode(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Info("Unicode: 🎉 🚀 ⭐")
		Infof("Emoji: %s %s", "😀", "🎊")
	})
}

func TestLogLevelChange(t *testing.T) {
	assert.NotPanics(t, func() {
		UpdateLogLevel("debug")
	})

	assert.NotPanics(t, func() {
		UpdateLogLevel("info")
	})

	assert.NotPanics(t, func() {
		UpdateLogLevel("warn")
	})

	assert.NotPanics(t, func() {
		UpdateLogLevel("error")
	})
}

func TestConcurrentLogging(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			Info("Concurrent log message")
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestConcurrentLogLevelUpdates(t *testing.T) {
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			UpdateLogLevel("debug")
			UpdateLogLevel("info")
			UpdateLogLevel("warn")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLoggingMixedLevels(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Debug("Debug message")
		Info("Info message")
		Warn("Warn message")
		Error("Error message")
		Debugf("Debug formatted")
		Infof("Info formatted")
		Warnf("Warn formatted")
		Errorf("Error formatted")
	})
}

func TestLogLevelUpdates(t *testing.T) {
	UpdateLogLevel("debug")
	Debug("This should be visible")

	UpdateLogLevel("info")
	Info("This should be visible")

	UpdateLogLevel("warn")
	Warn("This should be visible")

	UpdateLogLevel("error")
	Error("This should be visible")

	UpdateLogLevel("info")
}

func TestLoggingMethodsReturn(t *testing.T) {
	assert.NotPanics(t, func() {
		Debug("test")
	})

	assert.NotPanics(t, func() {
		Info("test")
	})

	assert.NotPanics(t, func() {
		Warn("test")
	})

	assert.NotPanics(t, func() {
		Error("test")
	})
}

func TestGetLoggerSingleton(t *testing.T) {
	logger1 := GetLogger()
	logger2 := GetLogger()
	logger3 := GetLogger()

	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)
	assert.NotNil(t, logger3)
	assert.Equal(t, logger1, logger2)
	assert.Equal(t, logger2, logger3)
	assert.Same(t, logger1, logger2)
}

func BenchmarkLoggingInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("Benchmark log message")
	}
}

func BenchmarkLoggingInfof(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Infof("Benchmark formatted log: %d", i)
	}
}

func BenchmarkUpdateLogLevel(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UpdateLogLevel("debug")
		UpdateLogLevel("info")
	}
}

func TestLoggerOutputFormat(t *testing.T) {
	assert.NotPanics(t, func() {
		Info("Test message")
		Infof("Test formatted %s", "message")
	})
}

func TestLoggerNilArgs(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		Infof("Message with nil: %v", nil)
	})
}

func TestLoggerVeryLongMessage(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	longMessage := strings.Repeat("a", 10000)

	assert.NotPanics(t, func() {
		Info(longMessage)
	})

	longFormat := strings.Repeat("%s ", 1000)
	args := make([]interface{}, 1000)
	for i := range args {
		args[i] = "test"
	}

	assert.NotPanics(t, func() {
		Infof(longFormat, args...)
	})
}

func TestLoggerManyLogs(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		for i := 0; i < 1000; i++ {
			Infof("Log message %d", i)
		}
	})
}
