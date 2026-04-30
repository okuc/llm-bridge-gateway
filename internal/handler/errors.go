package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// 错误码常量
const (
	ErrCodeInvalidRequest  = "INVALID_REQUEST"
	ErrCodeConversionError = "CONVERSION_ERROR"
	ErrCodeUpstreamError   = "UPSTREAM_ERROR"
	ErrCodeTimeoutError    = "TIMEOUT_ERROR"
	ErrCodeRateLimitError  = "RATE_LIMIT_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeInternalError   = "INTERNAL_ERROR"
)

// 错误响应辅助函数

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrCodeInvalidRequest,
			Message: message,
		},
	})
}

// ConversionError 422错误
func ConversionError(c *gin.Context, message string, details interface{}) {
	c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrCodeConversionError,
			Message: message,
			Details: details,
		},
	})
}

// UpstreamError 502错误
func UpstreamError(c *gin.Context, message string) {
	c.JSON(http.StatusBadGateway, ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrCodeUpstreamError,
			Message: message,
		},
	})
}

// TimeoutError 504错误
func TimeoutError(c *gin.Context, message string) {
	c.JSON(http.StatusGatewayTimeout, ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrCodeTimeoutError,
			Message: message,
		},
	})
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrCodeNotFound,
			Message: message,
		},
	})
}

// InternalError 500错误
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrCodeInternalError,
			Message: message,
		},
	})
}

// CustomError 自定义错误
func CustomError(c *gin.Context, statusCode int, code string, message string, details interface{}) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
