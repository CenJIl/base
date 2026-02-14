package web

import (
	"fmt"
)

// ErrUserNotFound 用户不存在错误
type ErrUserNotFound struct{}

func (e *ErrUserNotFound) Error() string {
	return "User not found"
}

// ErrUserExists 用户已存在错误
type ErrUserExists struct{}

func (e *ErrUserExists) Error() string {
	return "User already exists"
}

// ErrInvalidParam 无效参数错误
type ErrInvalidParam struct {
	Field string // 字段名
	Msg   string // 错误消息
}

func (e *ErrInvalidParam) Error() string {
	return fmt.Sprintf("Invalid parameter: %s", e.Field)
}

// ErrInternalError 内部错误
type ErrInternalError struct {
	Err error // 原始错误
}

func (e *ErrInternalError) Error() string {
	return fmt.Sprintf("Internal error: %v", e.Err)
}

// ErrDatabaseError 数据库错误
type ErrDatabaseError struct {
	Err error // 原始错误
	SQL string // SQL 语句
}

func (e *ErrDatabaseError) Error() string {
	return fmt.Sprintf("Database error: %v, SQL: %s", e.Err, e.SQL)
}
