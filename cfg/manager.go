package cfg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/CenJIl/base/common"
	"github.com/fsnotify/fsnotify"
)

const configFileName = "app.toml"

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrConfigInvalid  = errors.New("config file is invalid")
)

var (
	initOnce       sync.Once
	currentConfig  atomic.Pointer[any]
	changeHandlers []func(any)
	handlerMutex   sync.Mutex
	watcherMutex   sync.Mutex
	configWatcher  *fsnotify.Watcher
	cfgLog         common.Logger
)

// InitConfigWithLogger initializes app.toml beside the executable and watches it.
func InitConfigWithLogger[T any](defaultConfigRaw []byte, log common.Logger) {
	initOnce.Do(func() {
		if log == nil {
			log = &common.DefaultLog{}
		}
		cfgLog = log
		exePath, err := os.Executable()
		if err != nil {
			panic("获取可执行文件路径失败: " + err.Error())
		}
		configFilePath := filepath.Join(filepath.Dir(exePath), configFileName)
		data, err := os.ReadFile(configFilePath)
		if errors.Is(err, os.ErrNotExist) {
			cfgLog.Infof("配置文件不存在，写入默认配置: %s", configFilePath)
			if err := os.WriteFile(configFilePath, defaultConfigRaw, 0644); err != nil {
				panic("创建配置文件失败: " + err.Error())
			}
			data = defaultConfigRaw
		} else if err != nil {
			cfgLog.Errorf("读取配置文件失败，使用内存默认值: %v", err)
			data = defaultConfigRaw
		}
		var cfg T
		if err := toml.Unmarshal(data, &cfg); err != nil {
			if err := toml.Unmarshal(defaultConfigRaw, &cfg); err != nil {
				panic("配置初始化失败: " + err.Error())
			}
			cfgLog.Errorf("配置解析失败，使用内存默认值: %v", err)
		}
		storeConfig(&cfg)
		if err := startWatcher[T](configFilePath); err != nil {
			panic("创建配置监听失败: " + err.Error())
		}
	})
}

func InitConfig[T any](defaultConfigRaw []byte) {
	InitConfigWithLogger[T](defaultConfigRaw, &common.DefaultLog{})
}

// LoadConfig loads an existing TOML file and starts one replacement-aware watcher.
func LoadConfig[T any](configPath string) error {
	path, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("配置路径无效: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrConfigNotFound, path)
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	var cfg T
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigInvalid, err)
	}
	storeConfig(&cfg)
	if cfgLog == nil {
		cfgLog = &common.DefaultLog{}
	}
	return startWatcher[T](path)
}

func storeConfig[T any](cfg *T) {
	var value any = cfg
	currentConfig.Store(&value)
}

func startWatcher[T any](path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(filepath.Dir(path)); err != nil {
		_ = watcher.Close()
		return err
	}
	watcherMutex.Lock()
	old := configWatcher
	configWatcher = watcher
	watcherMutex.Unlock()
	if old != nil {
		_ = old.Close()
	}
	go watchConfig[T](watcher, filepath.Clean(path))
	return nil
}

// Close stops the active configuration watcher.
func Close() error {
	watcherMutex.Lock()
	watcher := configWatcher
	configWatcher = nil
	watcherMutex.Unlock()
	if watcher != nil {
		return watcher.Close()
	}
	return nil
}

// IsInitialized reports whether the requested configuration type is loaded.
func IsInitialized[T any]() bool {
	p := currentConfig.Load()
	if p == nil {
		return false
	}
	_, ok := (*p).(*T)
	return ok
}

// GetCfg returns nil before initialization or when the requested type differs.
func GetCfg[T any]() *T {
	p := currentConfig.Load()
	if p == nil {
		return nil
	}
	cfg, ok := (*p).(*T)
	if !ok {
		return nil
	}
	return cfg
}

func OnConfigChange[T any](h func(*T)) {
	if h == nil {
		return
	}
	handlerMutex.Lock()
	changeHandlers = append(changeHandlers, func(raw any) {
		if cfg, ok := raw.(*T); ok {
			h(cfg)
		}
	})
	handlerMutex.Unlock()
}

func watchConfig[T any](watcher *fsnotify.Watcher, configFilePath string) {
	var timer *time.Timer
	var timerMu sync.Mutex
	name := filepath.Base(configFilePath)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(filepath.Clean(event.Name)) != name || event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			timerMu.Lock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(100*time.Millisecond, func() {
				data, err := os.ReadFile(configFilePath)
				if err != nil {
					cfgLog.Errorf("配置热更新读取失败: %v", err)
					return
				}
				var cfg T
				if err := toml.Unmarshal(data, &cfg); err != nil {
					cfgLog.Errorf("配置热更新解析失败: %v", err)
					return
				}
				storeConfig(&cfg)
				handlerMutex.Lock()
				handlers := append([]func(any){}, changeHandlers...)
				handlerMutex.Unlock()
				var value any = &cfg
				for _, handler := range handlers {
					go handler(value)
				}
				cfgLog.Infof("配置已热更新")
			})
			timerMu.Unlock()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			cfgLog.Errorf("配置监听错误: %s", err)
		}
	}
}
