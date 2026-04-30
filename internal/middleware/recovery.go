package middleware

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
)

// Recovery Panic恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if isBrokenPipeError(se.Error()) {
							brokenPipe = true
						}
					}
				}
				if recoveredErr, ok := err.(error); ok && !brokenPipe {
					var se *os.SyscallError
					if errors.As(recoveredErr, &se) && isBrokenPipeError(se.Error()) {
						brokenPipe = true
					}
					var errno syscall.Errno
					if errors.As(recoveredErr, &errno) && (errno == syscall.EPIPE || errno == syscall.ECONNRESET) {
						brokenPipe = true
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					logger.Error("Recovery from panic",
						"error", err,
						"request", string(httpRequest),
					)
					c.Abort()
					return
				}

				logger.Error("Recovery from panic",
					"error", err,
					"request", string(httpRequest),
					"stack", string(debug.Stack()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Internal server error",
					},
				})
			}
		}()

		c.Next()
	}
}

func isBrokenPipeError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "broken pipe") || strings.Contains(lower, "connection reset by peer")
}
