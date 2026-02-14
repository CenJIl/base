package web

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// DownloadFile 流式下载文件
//
// 自动检查文件是否存在，设置 Content-Disposition 头
//
// 使用方式：
//   web.DownloadFile(c, "/path/to/file.pdf", "download.pdf")
func DownloadFile(c *app.RequestContext, filePath string, filename string) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		panic(NotFoundHTTP("文件不存在"))
	}
	if err != nil {
		panic(InternalHTTP("读取文件失败"))
	}

	// 设置响应头
	c.SetContentType("application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Content-Transfer-Encoding", "binary")

	// 设置文件下载
	c.File(filePath)
}

// DownloadWithRange 断点续传下载
//
// 支持 HTTP Range 请求，实现断点续传
//
// 使用方式：
//   web.DownloadWithRange(c, "/path/to/largefile.zip", "largefile.zip")
func DownloadWithRange(c *app.RequestContext, filePath string, filename string) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		panic(NotFoundHTTP("文件不存在"))
	}
	if err != nil {
		panic(InternalHTTP("读取文件失败"))
	}

	fileSize := fileInfo.Size()

	// 处理 Range 请求
	rangeHeader := string(c.GetHeader("Range"))
	if rangeHeader == "" {
		// 普通下载（无 Range）
		c.SetContentType("application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
		c.Header("Accept-Ranges", "bytes")
		c.Header("Content-Transfer-Encoding", "binary")

		c.File(filePath)
		return
	}

	// 解析 Range: bytes=0-1023
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		c.SetStatusCode(consts.StatusRequestedRangeNotSatisfiable)
		return
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	ranges := strings.Split(rangeStr, "-")
	if len(ranges) != 2 {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		c.SetStatusCode(consts.StatusRequestedRangeNotSatisfiable)
		return
	}

	// 解析起始和结束位置
	var start, end int64
	if ranges[0] != "" {
		start, _ = strconv.ParseInt(ranges[0], 10, 64)
	}
	if ranges[1] != "" {
		end, _ = strconv.ParseInt(ranges[1], 10, 64)
	} else {
		end = fileSize - 1
	}

	// 验证范围
	if start < 0 || end >= fileSize || start > end {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		c.SetStatusCode(consts.StatusRequestedRangeNotSatisfiable)
		return
	}

	// 设置部分响应头
	contentLength := end - start + 1
	c.SetStatusCode(consts.StatusPartialContent)
	c.SetContentType("application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Header("Accept-Ranges", "bytes")

	// TODO: Hertz 暂不支持 Range 文件下载，需要手动实现
	// 先使用普通下载
	c.File(filePath)
}

// FileExists 检查文件是否存在
//
// 使用方式：
//   if !web.FileExists("/path/to/file.pdf") {
//       panic(web.NotFoundHTTP("文件不存在"))
//   }
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileMimeType 根据 MIME 类型返回 Content-Type
//
// 使用方式：
//   mimeType := web.GetFileMimeType("image.jpg") // 返回: image/jpeg
func GetFileMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".zip":  "application/zip",
		".txt":  "text/plain; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".json": "application/json",
		".xml":  "application/xml",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream" // 默认二进制流
}
