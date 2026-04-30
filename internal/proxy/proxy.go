package proxy

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/okuc/llm-bridge-gateway/internal/config"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
)

// Proxy HTTP代理
type Proxy struct {
	client *http.Client
}

// New 创建代理实例
func New() *Proxy {
	return &Proxy{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Forward 转发请求到目标API
func (p *Proxy) Forward(req *http.Request, route *config.RouteConfig, body []byte) (*http.Response, error) {
	targetURL := route.Output.BaseURL

	proxyReq, err := http.NewRequestWithContext(req.Context(), req.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}

	proxyReq.Header.Set("Content-Type", "application/json")

	switch route.Output.Protocol {
	case config.ProtocolClaude:
		proxyReq.Header.Set("x-api-key", route.Output.APIKey)
		proxyReq.Header.Set("anthropic-version", "2023-06-01")
	case config.ProtocolOpenAI, config.ProtocolResponses:
		proxyReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", route.Output.APIKey))
	case config.ProtocolGemini:
		query := proxyReq.URL.Query()
		query.Set("key", route.Output.APIKey)
		proxyReq.URL.RawQuery = query.Encode()
	}

	if requestID := req.Header.Get("X-Request-ID"); requestID != "" {
		proxyReq.Header.Set("X-Request-ID", requestID)
	}

	resp, err := p.client.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}

	logger.Info("Request forwarded",
		"target", targetURL,
		"status", resp.StatusCode,
	)

	return resp, nil
}

// Close 关闭代理
func (p *Proxy) Close() {
	p.client.CloseIdleConnections()
}
