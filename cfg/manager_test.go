package cfg

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	AppName string `toml:"appName"`
	Port    int    `toml:"port"`
	Debug   bool   `toml:"debug"`
}

type MockLogger struct {
	logs   []string
	mu     sync.Mutex
	called bool
}

func (m *MockLogger) Infof(format string, v ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.logs = append(m.logs, "[INFO] "+format)
}

func (m *MockLogger) Errorf(format string, v ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.logs = append(m.logs, "[ERROR] "+format)
}

func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = nil
	m.called = false
}

func TestInitConfigWithLogger_Idempotent(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)
	config1 := GetCfg[TestConfig]()

	InitConfigWithLogger[TestConfig](defaultConfig, logger)
	config2 := GetCfg[TestConfig]()

	assert.Equal(t, config1, config2)
}

func TestInitConfig_DefaultValues(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	config := GetCfg[TestConfig]()
	assert.Equal(t, "TestApp", config.AppName)
	assert.Equal(t, 8080, config.Port)
	assert.True(t, config.Debug)
}

func TestInitConfig_Handler(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	OnConfigChange[TestConfig](func(cfg *TestConfig) {
		assert.NotNil(t, cfg)
	})

	config := GetCfg[TestConfig]()
	assert.NotNil(t, config)
	assert.NotNil(t, config)
}

func TestGetCfg_NilCheck(t *testing.T) {
	cfg := GetCfg[TestConfig]()
	if cfg == nil {
		t.Fatal("GetCfg should return non-nil pointer")
	}
}

func TestConfigTypes_Bool(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	config := GetCfg[TestConfig]()
	assert.True(t, config.Debug)
}

func TestConfigTypes_Number(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	config := GetCfg[TestConfig]()
	assert.Equal(t, 8080, config.Port)
}

func TestConfigTypes_String(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	config := GetCfg[TestConfig]()
	assert.Equal(t, "TestApp", config.AppName)
}

func TestInitConfig(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)

	InitConfig[TestConfig](defaultConfig)

	config := GetCfg[TestConfig]()
	assert.Equal(t, "TestApp", config.AppName)
	assert.Equal(t, 8080, config.Port)
	assert.True(t, config.Debug)
}

func TestOnConfigChange_MultipleHandlers(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	OnConfigChange[TestConfig](func(cfg *TestConfig) {
	})

	OnConfigChange[TestConfig](func(cfg *TestConfig) {
	})

	assert.NotNil(t, GetCfg[TestConfig]())
}

func TestConfigRaceCondition_ConcurrentGet(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			config := GetCfg[TestConfig]()
			assert.NotNil(t, config)
		}()
	}
	wg.Wait()
}

func TestConfigRaceCondition_ConcurrentHandlers(t *testing.T) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			OnConfigChange[TestConfig](func(cfg *TestConfig) {
			})
		}()
	}
	wg.Wait()

	assert.NotNil(t, GetCfg[TestConfig]())
}

func BenchmarkGetCfg(b *testing.B) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetCfg[TestConfig]()
	}
}

func BenchmarkOnConfigChange(b *testing.B) {
	defaultConfig := []byte(`appName = "TestApp"
port = 8080
debug = true`)
	logger := &MockLogger{}

	InitConfigWithLogger[TestConfig](defaultConfig, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OnConfigChange[TestConfig](func(cfg *TestConfig) {
		})
	}
}
