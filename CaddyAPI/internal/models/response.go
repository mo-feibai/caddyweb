package models

// API 响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// 错误响应
func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// 常用错误响应
func BadRequest(message string) *Response {
	return Error(400, message)
}

func Unauthorized(message string) *Response {
	return Error(401, message)
}

func NotFound(message string) *Response {
	return Error(404, message)
}

func InternalError(message string) *Response {
	return Error(500, message)
}
