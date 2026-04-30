package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查处理器
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "llm-bridge-gateway",
		})
	}
}
