package cfg

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	configFileName = "config.json"
)

var (
	currentConfig  atomic.Pointer[any]
	changeHandlers []func(any)
	handlerMutex   sync.Mutex
	cfgLog         CfgLogger
)

type CfgLogger interface {
	Info(msg string)
	Error(msg string)
}

type defaultLog struct{}

func (l *defaultLog) Info(msg string)  { log.Printf("INFO  %s", msg) }
func (l *defaultLog) Error(msg string) { log.Printf("ERROR %s", msg) }

// 使用go:embed 嵌入默认配置文件
func InitConfigWithLogger[T any](defaultConfigRaw []byte, log CfgLogger) error {
	cfgLog = log
	var cfg T

	exePath, _ := os.Executable()
	configFilePath := filepath.Join(filepath.Dir(exePath), configFileName)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		cfgLog.Info("配置文件不存在，写入默认配置")
		_ = os.WriteFile(configFilePath, defaultConfigRaw, 0644)
		if err := json.Unmarshal(defaultConfigRaw, &cfg); err != nil {
			return err
		}
	} else {
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			cfgLog.Error("读取配置文件失败，使用内存默认值")
			if err := json.Unmarshal(defaultConfigRaw, &cfg); err != nil {
				return err
			}
		} else if err := json.Unmarshal(data, &cfg); err != nil {
			cfgLog.Error("配置解析失败，使用内存默认值")
			_ = json.Unmarshal(defaultConfigRaw, &cfg)
		}
	}

	var anyCfg any = &cfg
	currentConfig.Store(&anyCfg)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go watchConfig[T](watcher, configFilePath)

	return watcher.Add(configFilePath)
}

// InitConfig 初始化配置，使用默认日志记录器
func InitConfig[T any](defaultConfigRaw []byte) error {
	return InitConfigWithLogger[T](defaultConfigRaw, &defaultLog{})
}

func watchConfig[T any](watcher *fsnotify.Watcher, configFilePath string) {
	var (
		timer    *time.Timer
		timerMu  sync.Mutex
		debounce = 100 * time.Millisecond
	)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write != fsnotify.Write {
				continue
			}

			timerMu.Lock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounce, func() {
				data, err := os.ReadFile(configFilePath)
				if err != nil {
					return
				}
				var cfg T
				if err := json.Unmarshal(data, &cfg); err != nil {
					cfgLog.Error("配置热更新解析失败")
					return
				}
				var anyCfg any = &cfg
				currentConfig.Store(&anyCfg)

				handlerMutex.Lock()
				for _, h := range changeHandlers {
					go h(anyCfg)
				}
				handlerMutex.Unlock()
				cfgLog.Info("配置已热更新")
			})
			timerMu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			cfgLog.Error("配置监听错误: " + err.Error())
		}
	}
}

func GetCfg[T any]() *T {
	p := currentConfig.Load()
	if p == nil {
		return new(T)
	}
	return (*p).(*T)
}

func OnConfigChange[T any](h func(cfg *T)) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	changeHandlers = append(changeHandlers, func(raw any) {
		h(raw.(*T))
	})
}
