package web

import "fmt"

// ErrorCode 业务错误码定义（类似 Spring Boot 的 HttpStatus）
type ErrorCode int

const (
	// 成功
	OK ErrorCode = 0

	// 客户端错误 (4xx) - 1xxxx
	BadRequest      ErrorCode = 10001 // 参数错误
	Unauthorized    ErrorCode = 10002 // 未授权
	Forbidden       ErrorCode = 10003 // 禁止访问
	NotFound        ErrorCode = 10004 // 资源不存在
	Conflict        ErrorCode = 10009 // 资源冲突
	TooManyRequests ErrorCode = 10020 // 请求过多

	// 业务逻辑错误 (2xxxx) - 可自定义
	UserNotFound ErrorCode = 20001 // 用户不存在
	UserExists   ErrorCode = 20002 // 用户已存在
	InvalidParam ErrorCode = 20003 // 无效参数

	// 服务端错误 (5xxxx) - 可自定义
	InternalError ErrorCode = 50001 // 内部错误
	DatabaseError ErrorCode = 50002 // 数据库错误
)

// ToHTTPStatus 转换为 HTTP 状态码
func (c ErrorCode) ToHTTPStatus() int {
	switch c / 100 {
	case 10:
		return 400 // Bad Request
	case 11:
		return 401 // Unauthorized
	case 12:
		return 403 // Forbidden
	case 14:
		return 404 // Not Found
	case 20:
		return 429 // Too Many Requests
	default:
		switch c / 1000 {
		case 1:
			return 500 // Internal Server Error
		case 2:
			return 503 // Service Unavailable
		default:
			return 500 // Internal Server Error
		}
	}
}

// Error 实现 error 接口
func (c ErrorCode) Error() string {
	switch c {
	case OK:
		return "success"
	case UserNotFound:
		return "User not found"
	case UserExists:
		return "User already exists"
	case InvalidParam:
		return "Invalid parameter"
	case InternalError, DatabaseError:
		return fmt.Sprintf("[%d] error", c)
	default:
		return fmt.Sprintf("[%d] error", c)
	}
}

// NewErr 创建新的错误
func NewErr(code ErrorCode) ErrorCode {
	return code
}
