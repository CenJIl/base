package cfg

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/CenJIl/base/common"
	"github.com/fsnotify/fsnotify"
)

const (
	configFileName = "config.toml"
)

var (
	initOnce       sync.Once
	currentConfig  atomic.Pointer[any]
	changeHandlers []func(any)
	handlerMutex   sync.Mutex
	cfgLog         common.Logger
)

// InitConfigWithLogger 使用指定的 Logger 初始化配置管理器
//
// 此函数会：
// 1. 在可执行文件所在目录查找 config.toml
// 2. 如果配置文件不存在，创建并写入默认配置
// 3. 如果配置文件存在，读取并解析
// 4. 解析失败时使用内存中的默认值并记录错误
// 5. 启动文件监听器，支持配置热更新
// 6. 使用 sync.Once 确保只初始化一次
//
// 参数
//
//	defaultConfigRaw - 默认配置的 TOML 格式字节数组
//	log - 自定义日志记录器，用于记录配置相关日志
//
// 注意事项
//   - 初始化失败会直接 panic，确保配置正确后再调用
//   - 配置文件名固定为 config.toml
//   - 热更新失败不会影响程序运行，仅记录错误
//   - 多次调用此函数，只有第一次生效（sync.Once 保证）
//
// 示例
//
//	type AppConfig struct {
//	    AppName string `toml:"appName"`
//	    Port    int    `toml:"port"`
//	    Debug   bool   `toml:"debug"`
//	}
//
//	defaultConfig := []byte(`appName = "MyApp"
//
// port = 8080
// debug = true`)
//
//	cfg.InitConfigWithLogger[AppConfig](defaultConfig, logger.GetLogger())
func InitConfigWithLogger[T any](defaultConfigRaw []byte, log common.Logger) {
	initOnce.Do(func() {
		cfgLog = log
		var cfg T

		exePath, err := os.Executable()
		if err != nil {
			panic("获取可执行文件路径失败: " + err.Error())
		}
		configFilePath := filepath.Join(filepath.Dir(exePath), configFileName)

		if _, err = os.Stat(configFilePath); os.IsNotExist(err) {
			cfgLog.Infof("配置文件不存在，写入默认配置")
			_ = os.WriteFile(configFilePath, defaultConfigRaw, 0644)
			if err = toml.Unmarshal(defaultConfigRaw, &cfg); err != nil {
				panic("配置初始化失败: " + err.Error())
			}
		} else {
			data, err := os.ReadFile(configFilePath)
			if err != nil {
				cfgLog.Errorf("读取配置文件失败，使用内存默认值")
				if err := toml.Unmarshal(defaultConfigRaw, &cfg); err != nil {
					panic("配置初始化失败: " + err.Error())
				}
			} else if err := toml.Unmarshal(data, &cfg); err != nil {
				cfgLog.Errorf("配置解析失败，使用内存默认值")
				_ = toml.Unmarshal(defaultConfigRaw, &cfg)
			}
		}

		var anyCfg any = &cfg
		currentConfig.Store(&anyCfg)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic("创建文件监听失败: " + err.Error())
		}

		if err := watcher.Add(configFilePath); err != nil {
			watcher.Close()
			panic("添加文件监听失败: " + err.Error())
		}

		go watchConfig[T](watcher, configFilePath)
	})
}

// InitConfig 使用默认日志记录器初始化配置管理器
//
// 这是 InitConfigWithLogger 的简化版本，内部使用 common.DefaultLog 作为日志记录器
// 适用于不需要自定义日志记录器的场景
//
// 参数
//
//	defaultConfigRaw - 默认配置的 TOML 格式字节数组
//
// 示例
//
//	cfg.InitConfig[AppConfig](defaultConfig)
func InitConfig[T any](defaultConfigRaw []byte) {
	InitConfigWithLogger[T](defaultConfigRaw, &common.DefaultLog{})
}

// GetCfg 获取当前配置的指针
//
// 返回当前配置的只读指针。如果配置未初始化，返回零值指针。
// 此方法是线程安全的，可以在任何 goroutine 中安全调用
//
// 返回值
//
//	*T - 配置结构体的指针，类型由泛型参数 T 决定
//
// 注意事项
//   - 返回的指针指向配置的当前状态，配置热更新后需要重新调用获取最新值
//   - 如果未调用 InitConfig 初始化，返回零值指针
//   - 此方法不复制配置，直接返回内部指针，不要修改返回值的内容
//
// 示例
//
//	config := cfg.GetCfg[AppConfig]()
//	fmt.Printf("AppName: %s\n", config.AppName)
func GetCfg[T any]() *T {
	p := currentConfig.Load()
	if p == nil {
		return new(T)
	}
	return (*p).(*T)
}

// OnConfigChange 注册配置变更回调函数
//
// 当配置文件被修改并重新加载成功后，所有注册的回调函数会被调用。
// 回调函数在独立的 goroutine 中异步执行，确保不影响主流程。
// 多次调用此函数可以注册多个回调，配置变更时所有回调都会被触发
//
// 参数
//
//	h - 配置变更时的回调函数，接收新配置的指针
//
// 注意事项
//   - 回调函数在独立 goroutine 中执行，需要自行处理并发安全
//   - 回调函数中不应执行耗时操作，避免阻塞
//   - 回调函数执行失败不会影响其他回调
//   - 配置解析失败时，回调函数不会被调用
//
// 示例
//
//	cfg.OnConfigChange(func(newCfg *AppConfig) {
//	    logger.Infof("配置已更新: %+v", newCfg)
//	    // 执行配置变更后的逻辑
//	})
func OnConfigChange[T any](h func(cfg *T)) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	changeHandlers = append(changeHandlers, func(raw any) {
		h(raw.(*T))
	})
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
				if err := toml.Unmarshal(data, &cfg); err != nil {
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
