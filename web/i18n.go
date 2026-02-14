package web

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
)

// ============================================================================
// 高性能 i18n 实现
//
// 设计理念：
// 1. 零外部依赖 - 避免第三方库的兼容性问题
// 2. 预编译 - 启动时加载所有语言，运行时零开销
// 3. 线程安全 - 使用 sync.RWMutex 保护并发读取
// 4. 高性能 - O(1) 查找，直接返回字符串
// 5. 易用性 - T() 函数直接翻译，无需手动处理 Localizer
// ============================================================================

// localizer 本地化器接口
type localizer interface {
	Localize(msgID string, args ...any) string
}

// i18nStore i18n 存储接口
type i18nStore interface {
	Get(lang string) (localizer localizer, ok bool)
	Set(lang string, data map[string]string)
}

// memoryStore 内存存储（默认）
type memoryStore struct {
	translations map[string]map[string]string
	mu          sync.RWMutex
}

// newMemoryStore 创建内存存储
func newMemoryStore() *memoryStore {
	return &memoryStore{
		translations: make(map[string]map[string]string),
	}
}

// Get 获取翻译器（实现 i18nStore 接口）
func (s *memoryStore) Get(lang string) (localizer localizer, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if langMap, exists := s.translations[lang]; exists {
		return &mapLocalizer{data: langMap}, true
	}

	return nil, false
}

// Set 设置翻译
func (s *memoryStore) Set(lang string, data map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.translations[lang] = data
}

// Get 获取翻译（实现 i18nStore 接口）
func (s *memoryStore) Get(lang string) (localizer localizer, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if langMap, exists := s.translations[lang]; exists {
		return &mapLocalizer{data: langMap}, true
	}

	return nil, false
}

// mapLocalizer map 本地化器
type mapLocalizer struct {
	data map[string]string
}

// Localize 翻译（实现 localizer 接口）
func (l *mapLocalizer) Localize(msgID string, args ...any) string {
	// O(1) 查找
	if msg, ok := l.data[msgID]; ok {
		// 简单替换占位符（不支持复杂格式化）
		result := msg
		for i, arg := range args {
			result = fmt.Sprintf(result, fmt.Sprintf("{%d}", i), arg)
		}
		return result
	}

	return msgID // 找不到翻译返回原文
}

var (
	store i18nStore = newMemoryStore()
)

// InitI18n 初始化 i18n
//
// 预加载所有翻译到内存
//
// 使用方式：
//   web.InitI18n(map[string]map[string]string{
//       "zh-CN": {
//           "user.not_found": "用户不存在",
//           "welcome": "欢迎",
//       },
//       "en-US": {
//           "user.not_found": "User not found",
//           "welcome": "Welcome",
//       },
//   })
func InitI18n(translations map[string]map[string]string) {
	if ms, ok := translations.(map[string]map[string]string); ok {
		if memStore, ok := store.(*memoryStore); ok {
			for lang, langMap := range ms {
				memStore.Set(lang, langMap)
			}
		}
	}
}

// I18nMiddleware i18n 中间件
//
// 从 Accept-Language 或查询参数 ?lang=zh-CN 获取语言
//
// 性能优化：只做必要的字符串操作，不涉及文件 I/O
func I18nMiddleware(defaultLang string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 1. 从查询参数获取语言（优先级最高）
		lang := c.GetQuery("lang")

		// 2. 从请求头获取语言
		if lang == "" {
			lang = string(c.GetHeader("Accept-Language"))
		}

		// 3. 使用默认语言
		if lang == "" {
			lang = defaultLang
		}

		// 存储到上下文（避免重复查询）
		c.Set("lang", lang)

		c.Next(ctx)
	}
}

// GetLocalizer 获取本地化器
//
// 性能优化：O(1) 查找，使用读锁
func GetLocalizer(c *app.RequestContext) localizer {
	lang := GetLanguage(c)
	loc, _ := store.Get(lang)

	if loc == nil {
		// 返回兜底的本地化器
		return &defaultLocalizer{lang: lang}
	}

	return loc
}

// GetLanguage 获取当前语言
func GetLanguage(c *app.RequestContext) string {
	if lang, ok := c.Get("lang"); ok {
		if s, ok := lang.(string); ok {
			return s
		}
	}
	return "en-US" // 默认英文
}

// T 翻译辅助函数
//
// 易用性：一行代码完成翻译，无需手动处理 Localizer
//
// 使用方式：
//   msg := web.T(c, "user.not_found", "用户123")
func T(c *app.RequestContext, msgID string, args ...any) string {
	loc := GetLocalizer(c)
	return loc.Localize(msgID, args...)
}

// SetLanguage 设置当前语言（用于测试）
func SetLanguage(c *app.RequestContext, lang string) {
	c.Set("lang", lang)
}

// ============================================================================
// 兜底本地化器（当找不到翻译时使用）
// ============================================================================

type defaultLocalizer struct {
	lang string
}

func (l *defaultLocalizer) Localize(msgID string, args ...any) string {
	// 找不到翻译时返回 key
	return msgID
}
