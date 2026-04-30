package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/internal/converter"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
	opentrans "github.com/xy200303/OpenTrans"
)

// StreamHandler 流式响应处理器
type StreamHandler struct {
	converter *converter.ProtocolConverter
}

// NewStreamHandler 创建流式处理器
func NewStreamHandler(converter *converter.ProtocolConverter) *StreamHandler {
	return &StreamHandler{
		converter: converter,
	}
}

// HandleStream 处理流式响应
func (sh *StreamHandler) HandleStream(c *gin.Context, resp *http.Response, source, target opentrans.Protocol) {
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"code":    "UPSTREAM_ERROR",
					"message": "Failed to read upstream error response",
				},
			})
			return
		}
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/json"
		}
		c.Data(resp.StatusCode, contentType, body)
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 获取flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "STREAM_ERROR",
				"message": "Streaming not supported",
			},
		})
		return
	}

	// 读取并转发SSE事件
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if line == "" {
			fmt.Fprintf(c.Writer, "\n")
			flusher.Flush()
			continue
		}

		// 处理SSE数据
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查结束标记
			if data == "[DONE]" {
				fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				flusher.Flush()
				break
			}

			// 转换事件
			converted, err := sh.converter.ConvertStreamEvent([]byte(data), source, target)
			if err != nil {
				logger.Error("Stream conversion error", "error", err)
				continue
			}

			fmt.Fprintf(c.Writer, "data: %s\n\n", converted)
			flusher.Flush()
		}
	}

	// 检查扫描错误
	if err := scanner.Err(); err != nil {
		logger.Error("Stream scan error", "error", err)
	}
}

// IsStreamResponse 检查是否是流式响应
func IsStreamResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "text/event-stream")
}

// CopyHeaders 复制响应头
func CopyHeaders(dst http.ResponseWriter, src *http.Response) {
	for key, values := range src.Header {
		for _, value := range values {
			dst.Header().Add(key, value)
		}
	}
}
