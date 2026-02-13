package middleware

import (
	"github.com/google/uuid"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
	TraceID string      `json:"traceId"`
}

func NewResponse(code int, msg string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: msg,
		Data:    data,
		TraceID: uuid.New().String(),
	}
}

func Success(data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
		TraceID: uuid.New().String(),
	}
}

func Error(code int, msg string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: msg,
		Data:    data,
		TraceID: uuid.New().String(),
	}
}
