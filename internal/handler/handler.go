package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/internal/config"
	"github.com/okuc/llm-bridge-gateway/internal/converter"
	"github.com/okuc/llm-bridge-gateway/internal/proxy"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
	opentrans "github.com/xy200303/OpenTrans"
)

// GatewayHandler 网关处理器
type GatewayHandler struct {
	converter     *converter.ProtocolConverter
	proxy         *proxy.Proxy
	route         *config.RouteConfig
	streamHandler *proxy.StreamHandler
}

// NewGatewayHandler 创建网关处理器
func NewGatewayHandler(
	converter *converter.ProtocolConverter,
	proxy *proxy.Proxy,
	route *config.RouteConfig,
	streamHandler *proxy.StreamHandler,
) *GatewayHandler {
	return &GatewayHandler{
		converter:     converter,
		proxy:         proxy,
		route:         route,
		streamHandler: streamHandler,
	}
}

// HandleRequest 处理请求
func (h *GatewayHandler) HandleRequest(c *gin.Context) {
	// 1. 限制请求体大小（10MB）并读取
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		BadRequest(c, "Failed to read request body (max 10MB)")
		return
	}
	defer c.Request.Body.Close()

	// 检查空请求体
	if len(body) == 0 {
		BadRequest(c, "Request body is empty")
		return
	}

	logger.Info("Received request",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"content_length", len(body),
	)

	// 2. 获取协议类型
	sourceProtocol, err := converter.GetProtocol(h.route.Input.Protocol)
	if err != nil {
		ConversionError(c, "Invalid input protocol", err.Error())
		return
	}

	targetProtocol, err := converter.GetProtocol(h.route.Output.Protocol)
	if err != nil {
		ConversionError(c, "Invalid output protocol", err.Error())
		return
	}

	// 3. 检查是否是流式请求
	isStream := converter.IsStreamRequest(body)

	// 4. 转换请求体
	convertedBody, err := h.converter.ConvertRequest(body, sourceProtocol, targetProtocol)
	if err != nil {
		logger.Error("Request conversion failed",
			"error", err,
			"source", sourceProtocol,
			"target", targetProtocol,
		)
		ConversionError(c, "Failed to convert request body", err.Error())
		return
	}

	logger.Debug("Request converted",
		"source", sourceProtocol,
		"target", targetProtocol,
		"original_size", len(body),
		"converted_size", len(convertedBody),
	)

	// 5. 转发请求（应用路由超时）
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.route.Timeout)
	defer cancel()

	proxyReq := c.Request.Clone(ctx)
	resp, err := proxy.WithRetry(ctx, h.route, func() (*http.Response, error) {
		return h.proxy.Forward(proxyReq, h.route, convertedBody)
	})
	if err != nil {
		logger.Error("Request forwarding failed", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			TimeoutError(c, "Upstream request timed out")
			return
		}
		UpstreamError(c, "Failed to forward request to upstream API")
		return
	}
	defer resp.Body.Close()

	// 6. 处理响应
	if isStream && proxy.IsStreamResponse(resp) {
		// 流式响应
		h.handleStreamResponse(c, resp, sourceProtocol, targetProtocol)
	} else {
		// 非流式响应
		h.handleSyncResponse(c, resp, sourceProtocol, targetProtocol)
	}
}

// handleSyncResponse 处理同步响应
func (h *GatewayHandler) handleSyncResponse(c *gin.Context, resp *http.Response, source, target opentrans.Protocol) {
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		UpstreamError(c, "Failed to read upstream response")
		return
	}

	// 如果状态码不是2xx，直接返回错误
	if resp.StatusCode >= 400 {
		c.Data(resp.StatusCode, "application/json", body)
		return
	}

	// 转换响应体
	// 注意：响应转换的方向与请求相反
	// 请求：source -> target（客户端发送source格式，目标API期望target格式）
	// 响应：target -> source（目标API返回target格式，客户端期望source格式）
	convertedBody, err := h.converter.ConvertResponse(body, target, source)
	if err != nil {
		logger.Error("Response conversion failed",
			"error", err,
			"source", target,
			"target", source,
		)
		ConversionError(c, "Failed to convert response body", err.Error())
		return
	}

	logger.Info("Response converted",
		"source", target,
		"target", source,
		"original_size", len(body),
		"converted_size", len(convertedBody),
	)

	// 返回响应
	c.Data(resp.StatusCode, "application/json", convertedBody)
}

// handleStreamResponse 处理流式响应
func (h *GatewayHandler) handleStreamResponse(c *gin.Context, resp *http.Response, source, target opentrans.Protocol) {
	// 注意：流式响应转换的方向与请求相反
	// 请求：source -> target
	// 响应：target -> source
	h.streamHandler.HandleStream(c, resp, target, source)
}
