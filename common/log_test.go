package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLog_Infof(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Info message")
	})

	log.Infof("Test info message")
}

func TestDefaultLog_Errorf(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Errorf("Error message")
	})

	log.Errorf("Test error message")
}

func TestDefaultLog_Infof_WithArgs(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Info with %s", "args")
	})

	assert.NotPanics(t, func() {
		log.Infof("Info with %d %s %v", 42, "test", true)
	})

	log.Infof("Test with args: %s %d", "value", 100)
}

func TestDefaultLog_Errorf_WithArgs(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Errorf("Error with %s", "args")
	})

	assert.NotPanics(t, func() {
		log.Errorf("Error with %d %s %v", 42, "test", false)
	})

	log.Errorf("Error with args: %s %d", "value", 200)
}

func TestDefaultLog_NoArgs(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("No args")
	})

	assert.NotPanics(t, func() {
		log.Errorf("No args error")
	})

	log.Infof("Message without arguments")
	log.Errorf("Error without arguments")
}

func TestDefaultLog_NilArgs(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Nil arg: %v", nil)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Nil arg error: %v", nil)
	})
}

func TestDefaultLog_EmptyFormat(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("")
	})

	assert.NotPanics(t, func() {
		log.Errorf("")
	})
}

func TestDefaultLog_SpecialCharacters(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Special chars: %s %d %v", "test", 42, true)
	})

	assert.NotPanics(t, func() {
		log.Infof("Symbols: !@#$%^&*()")
	})

	log.Infof("Special characters: %s", "test!@#$%")
	log.Errorf("Error symbols: %% %d %s", 100, "test")
}

func TestDefaultLog_Newlines(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Line 1\nLine 2\nLine 3")
	})

	assert.NotPanics(t, func() {
		log.Errorf("Error\n\nMultiple\n\nLines")
	})

	log.Infof("Multi-line\nmessage")
	log.Errorf("Multi-line\nerror\nmessage")
}

func TestDefaultLog_Tabs(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Tab\tseparated\tvalues")
	})

	assert.NotPanics(t, func() {
		log.Errorf("Error\ttab\tmessage")
	})
}

func TestDefaultLog_ChineseCharacters(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("中文日志消息")
	})

	assert.NotPanics(t, func() {
		log.Infof("格式化中文: %s", "测试")
	})

	assert.NotPanics(t, func() {
		log.Errorf("中文错误消息: %s", "错误")
	})

	log.Infof("中文测试: %s %d", "值", 100)
	log.Errorf("中文错误: %v", "测试错误")
}

func TestDefaultLog_Unicode(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Unicode: 🎉 🚀 ⭐")
	})

	assert.NotPanics(t, func() {
		log.Errorf("Emoji: %s %s", "😀", "🎊")
	})
}

func TestDefaultLog_LongMessage(t *testing.T) {
	log := &DefaultLog{}
	longMessage := strings.Repeat("a", 1000)

	assert.NotPanics(t, func() {
		log.Infof(longMessage)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Error: %s", longMessage)
	})
}

func TestDefaultLog_ManyArgs(t *testing.T) {
	log := &DefaultLog{}

	args := make([]interface{}, 10)
	for i := range args {
		args[i] = i
	}

	assert.NotPanics(t, func() {
		log.Infof("Many args: %v %v %v %v %v %v %v %v %v %v", args...)
	})
}

func TestDefaultLog_Verbs(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("String: %s, Number: %d, Boolean: %v", "test", 42, true)
	})

	assert.NotPanics(t, func() {
		log.Infof("Float: %f, Hex: %x", 3.14, 255)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Error: %s, Code: %d, Success: %v", "test", 500, false)
	})
}

func TestDefaultLog_VeryLongFormat(t *testing.T) {
	log := &DefaultLog{}
	longFormat := strings.Repeat("%s ", 100)
	args := make([]interface{}, 100)
	for i := range args {
		args[i] = "test"
	}

	assert.NotPanics(t, func() {
		log.Infof(longFormat, args...)
	})
}

func TestDefaultLog_Interface(t *testing.T) {
	var logger Logger = &DefaultLog{}
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		logger.Infof("Interface info")
	})

	assert.NotPanics(t, func() {
		logger.Errorf("Interface error")
	})

	logger.Infof("Interface test: %s", "message")
	logger.Errorf("Interface error: %d", 404)
}

func TestDefaultLog_MultipleLogs(t *testing.T) {
	log := &DefaultLog{}

	for i := 0; i < 100; i++ {
		assert.NotPanics(t, func() {
			log.Infof("Log message %d", i)
		})
	}

	for i := 0; i < 100; i++ {
		assert.NotPanics(t, func() {
			log.Errorf("Error message %d", i)
		})
	}
}

func TestDefaultLog_Concurrent(t *testing.T) {
	log := &DefaultLog{}
	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(n int) {
			log.Infof("Concurrent info %d", n)
			log.Errorf("Concurrent error %d", n)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestDefaultLog_PointerArg(t *testing.T) {
	log := &DefaultLog{}
	value := 42

	assert.NotPanics(t, func() {
		log.Infof("Pointer: %p", &value)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Pointer error: %p", &value)
	})
}

func TestDefaultLog_StructArg(t *testing.T) {
	log := &DefaultLog{}

	type TestStruct struct {
		Name  string
		Value int
	}

	s := TestStruct{Name: "test", Value: 100}

	assert.NotPanics(t, func() {
		log.Infof("Struct: %v", s)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Struct error: %+v", s)
	})
}

func TestDefaultLog_SliceArg(t *testing.T) {
	log := &DefaultLog{}
	slice := []int{1, 2, 3, 4, 5}

	assert.NotPanics(t, func() {
		log.Infof("Slice: %v", slice)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Slice error: %v", slice)
	})
}

func TestDefaultLog_MapArg(t *testing.T) {
	log := &DefaultLog{}
	m := map[string]int{"a": 1, "b": 2}

	assert.NotPanics(t, func() {
		log.Infof("Map: %v", m)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Map error: %v", m)
	})
}

func TestDefaultLog_PercentSign(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Literal %% percent")
	})

	assert.NotPanics(t, func() {
		log.Errorf("Literal %% percent error")
	})

	log.Infof("This is a %% literal")
	log.Errorf("Error with %% literal")
}

func TestDefaultLog_RetainFormat(t *testing.T) {
	log := &DefaultLog{}

	assert.NotPanics(t, func() {
		log.Infof("Value: %d", 42)
	})

	assert.NotPanics(t, func() {
		log.Errorf("Error code: %d", 500)
	})
}

func BenchmarkDefaultLog_Infof(b *testing.B) {
	log := &DefaultLog{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Infof("Benchmark log message %d", i)
	}
}

func BenchmarkDefaultLog_Errorf(b *testing.B) {
	log := &DefaultLog{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Errorf("Benchmark error message %d", i)
	}
}
