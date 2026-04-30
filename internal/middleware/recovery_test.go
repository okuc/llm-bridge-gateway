package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
)

func TestRecoveryHandlesNonErrorPanic(t *testing.T) {
	logger.Init("error", "text", "stdout", "")
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(Recovery())
	engine.GET("/panic", func(c *gin.Context) {
		panic("plain string panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
