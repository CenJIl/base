package cfg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/CenJIl/base/common"
	"github.com/fsnotify/fsnotify"
)

const (
	configFileName = "config.json"
)

var (
	initOnce       sync.Once
	currentConfig  atomic.Pointer[any]
	changeHandlers []func(any)
	handlerMutex   sync.Mutex
	cfgLog         common.Logger
)

// InitConfigWithLogger 初始化配置，失败直接 panic，使用 sync.Once 保证只初始化一次
func InitConfigWithLogger[T any](defaultConfigRaw []byte, log common.Logger) {
	initOnce.Do(func() {
		cfgLog = log
		var cfg T

		exePath, err := os.Executable()
		if err != nil {
			panic("获取可执行文件路径失败: " + err.Error())
		}
		configFilePath := filepath.Join(filepath.Dir(exePath), configFileName)

		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			cfgLog.Infof("配置文件不存在，写入默认配置")
			_ = os.WriteFile(configFilePath, defaultConfigRaw, 0644)
			if err := json.Unmarshal(defaultConfigRaw, &cfg); err != nil {
				panic("配置初始化失败: " + err.Error())
			}
		} else {
			data, err := os.ReadFile(configFilePath)
			if err != nil {
				cfgLog.Errorf("读取配置文件失败，使用内存默认值")
				if err := json.Unmarshal(defaultConfigRaw, &cfg); err != nil {
					panic("配置初始化失败: " + err.Error())
				}
			} else if err := json.Unmarshal(data, &cfg); err != nil {
				cfgLog.Errorf("配置解析失败，使用内存默认值")
				_ = json.Unmarshal(defaultConfigRaw, &cfg)
			}
		}

		var anyCfg any = &cfg
		currentConfig.Store(&anyCfg)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic("创建文件监听失败: " + err.Error())
		}

		// 先 Add 文件，成功后再启动 goroutine，保证原子性
		if err := watcher.Add(configFilePath); err != nil {
			watcher.Close()
			panic("添加文件监听失败: " + err.Error())
		}

		go watchConfig[T](watcher, configFilePath)
	})
}

// InitConfig 初始化配置，使用默认日志记录器
func InitConfig[T any](defaultConfigRaw []byte) {
	InitConfigWithLogger[T](defaultConfigRaw, &common.DefaultLog{})
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
					cfgLog.Errorf("配置热更新解析失败")
					return
				}
				var anyCfg any = &cfg
				currentConfig.Store(&anyCfg)

				handlerMutex.Lock()
				for _, h := range changeHandlers {
					go h(anyCfg)
				}
				handlerMutex.Unlock()
				cfgLog.Infof("配置已热更新")
			})
			timerMu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			cfgLog.Errorf("配置监听错误: %s", err.Error())
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
