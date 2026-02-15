package web

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// UploadConfig 上传配置
type UploadConfig struct {
	MaxFileSize int64    `toml:"maxFileSize"` // 单文件最大大小（字节）
	AllowedExts []string `toml:"allowedExts"` // 允许的扩展名（如 [".jpg", ".png"]）
	UploadPath  string   `toml:"uploadPath"`  // 上传保存路径
	URLPrefix   string   `toml:"urlPrefix"`   // 访问 URL 前缀
}

// UploadMiddleware 上传中间件（验证 + 限制）
//
// 检查 Content-Type 并将配置保存到上下文，供 handler 使用
//
// 使用方式：
//
//	h.POST("/upload", web.UploadMiddleware(config.Upload), uploadHandler)
func UploadMiddleware(config UploadConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 检查 Content-Type
		contentType := string(c.GetHeader("Content-Type"))
		if !strings.Contains(contentType, "multipart/form-data") {
			panic(BadRequestHTTP("必须是 multipart/form-data 格式"))
		}

		// 保存配置到上下文（handler 中使用）
		c.Set("upload_config", config)
		c.Next(ctx)
	}
}

// SaveUploadedFile 保存上传文件到指定路径
//
// 自动创建父目录
//
// 使用方式：
//
//	file, _ := c.FormFile("file")
//	err := web.SaveUploadedFile(file, "/path/to/save/filename.ext")
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, src)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// IsAllowedExt 检查文件扩展名是否允许
//
// 不区分大小写
//
// 使用方式：
//
//	allowed := []string{".jpg", ".png", ".pdf"}
//	if !web.IsAllowedExt("image.JPG", allowed) {
//	    return "不支持的文件类型"
//	}
func IsAllowedExt(filename string, allowedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExts {
		if ext == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// ValidateFile 验证文件（大小 + 扩展名）
//
// 返回错误信息，验证通过返回 nil
//
// 使用方式：
//
//	file, _ := c.FormFile("file")
//	if err := web.ValidateFile(file, config.Upload); err != nil {
//	    panic(web.BadRequestHTTP(err.Error()))
//	}
func ValidateFile(file *multipart.FileHeader, config UploadConfig) error {
	// 检查大小
	if file.Size > config.MaxFileSize {
		return fmt.Errorf("文件大小超限：%.2f MB / %.2f MB",
			float64(file.Size)/1024/1024,
			float64(config.MaxFileSize)/1024/1024)
	}

	// 检查扩展名
	if len(config.AllowedExts) > 0 && !IsAllowedExt(file.Filename, config.AllowedExts) {
		return fmt.Errorf("不支持的文件类型：%s（允许：%s）",
			filepath.Ext(file.Filename),
			strings.Join(config.AllowedExts, ", "))
	}

	return nil
}

// GetUploadConfig 从上下文获取上传配置
//
// 使用方式：
//
//	config := web.GetUploadConfig(c)
//	maxsize := config.MaxFileSize
func GetUploadConfig(c *app.RequestContext) UploadConfig {
	v, _ := c.Get("upload_config")
	if v != nil {
		if config, ok := v.(UploadConfig); ok {
			return config
		}
	}
	return UploadConfig{} // 返回零值
}

// GenerateFilename 生成安全的文件名
//
// 使用时间戳 + 原始扩展名，避免文件名冲突
//
// 使用方式：
//
//	filename := web.GenerateFilename("photo.jpg") // 返回: 20240215123456.jpg
func GenerateFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	// 移除扩展名中的特殊字符
	ext = strings.TrimPrefix(ext, ".")
	return fmt.Sprintf("%d.%s", time.Now().Unix(), ext)
}
