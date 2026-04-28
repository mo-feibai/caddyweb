package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一 API 响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// 成功响应（带自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// 错误响应（不暴露内部错误详情）
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// 错误响应（带内部错误，用于日志记录但不返回给前端）
func ErrorWithInternalLog(c *gin.Context, httpStatus int, code int, message string, internalErr error) {
	if internalErr != nil {
		log.Printf("[ERROR] %s | Internal Error: %v", message, internalErr)
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// 常用错误码
const (
	CodeSuccess        = 0
	CodeBadRequest     = 400
	CodeUnauthorized   = 401
	CodeForbidden      = 403
	CodeNotFound       = 404
	CodeInternalError  = 500
	CodeServiceUnavail = 503
)

// 常用错误响应（不记录日志）
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeInternalError, message)
}

func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, CodeServiceUnavail, message)
}

// 常用错误响应（记录内部错误到日志）
func BadRequestWithLog(c *gin.Context, message string, err error) {
	ErrorWithInternalLog(c, http.StatusBadRequest, CodeBadRequest, message, err)
}

func InternalServerErrorWithLog(c *gin.Context, message string, err error) {
	ErrorWithInternalLog(c, http.StatusInternalServerError, CodeInternalError, message, err)
}

func ServiceUnavailableWithLog(c *gin.Context, message string, err error) {
	ErrorWithInternalLog(c, http.StatusServiceUnavailable, CodeServiceUnavail, message, err)
}
