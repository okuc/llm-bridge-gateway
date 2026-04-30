package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		duration := time.Since(start)

		// 获取请求信息
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		userAgent := c.Request.UserAgent()

		// 获取响应大小
		bodySize := c.Writer.Size()

		// 构建日志字段
		args := []any{
			"status", statusCode,
			"method", method,
			"path", path,
			"query", query,
			"ip", clientIP,
			"user_agent", userAgent,
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"body_size", bodySize,
		}

		// 添加请求ID（如果有）
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			args = append(args, "request_id", requestID)
		}

		// 添加错误信息（如果有）
		if len(c.Errors) > 0 {
			args = append(args, "errors", c.Errors.String())
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error("Request completed", args...)
		case statusCode >= 400:
			logger.Warn("Request completed", args...)
		default:
			logger.Info("Request completed", args...)
		}
	}
}
