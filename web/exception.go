package web

import (
	"fmt"
)

// Exception 业务异常基类（类似 Spring Boot 的 Exception）
type Exception struct {
	Code    int    // 业务码
	Message string // 错误消息
}

// Error 实现 error 接口
func (e *Exception) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// HTTPException HTTP 异常（带有 HTTP 状态码）
type HTTPException struct {
	HTTPStatus int    // HTTP 状态码
	Code       int    // 业务码
	Message    string // 错误消息
}

// Error 实现 error 接口
func (e *HTTPException) Error() string {
	return fmt.Sprintf("[HTTP %d] [%d] %s", e.HTTPStatus, e.Code, e.Message)
}

// NewException 创建业务异常
func NewException(code int, message string) *Exception {
	return &Exception{Code: code, Message: message}
}

// NewHTTPException 创建 HTTP 异常
func NewHTTPException(httpStatus int, code int, message string) *HTTPException {
	return &HTTPException{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}

// BadRequestHTTP 400 错误
func BadRequestHTTP(msg string) *HTTPException {
	return NewHTTPException(400, 400, msg)
}

// UnauthorizedHTTP 401 错误
func UnauthorizedHTTP(msg string) *HTTPException {
	return NewHTTPException(401, 401, msg)
}

// ForbiddenHTTP 403 错误
func ForbiddenHTTP(msg string) *HTTPException {
	return NewHTTPException(403, 403, msg)
}

// NotFoundHTTP 404 错误
func NotFoundHTTP(msg string) *HTTPException {
	return NewHTTPException(404, 404, msg)
}

// ConflictHTTP 409 错误
func ConflictHTTP(msg string) *HTTPException {
	return NewHTTPException(409, 409, msg)
}

// InternalHTTP 500 错误
func InternalHTTP(msg string) *HTTPException {
	return NewHTTPException(500, 500, msg)
}
