package web

// Result 统一响应结构（类似 Spring Boot 的 Result<T>）
type Result struct {
	Code    int    `json:"code"`              // 业务码：0=成功，其他=错误
	Message string `json:"message"`           // 响应消息
	Data    any    `json:"data"`              // 响应数据
	TraceID string `json:"traceId,omitempty"` // 链路追踪 ID（由 WrapHandler 或用户设置）
}

// PagedData 分页数据
type PagedData struct {
	Items     any   `json:"items"`     // 数据列表
	Page      int   `json:"page"`      // 当前页码
	PageSize  int   `json:"pageSize"`  // 每页大小
	Total     int64 `json:"total"`     // 总记录数
	TotalPage int   `json:"totalPage"` // 总页数
}

// Success 成功响应
func Success(data any) Result {
	return Result{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// Fail 失败响应
func Fail(code int, message string) Result {
	return Result{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// FailWithData 失败响应（带自定义数据）
func FailWithData(code int, message string, data any) Result {
	return Result{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// PagedSuccess 分页成功响应
func PagedSuccess(items any, page, pageSize int, total int64) Result {
	totalPage := int(total) / pageSize
	if total%int64(pageSize) != 0 {
		totalPage++
	}

	return Result{
		Code:    0,
		Message: "success",
		Data: PagedData{
			Items:     items,
			Page:      page,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPage,
		},
	}
}
