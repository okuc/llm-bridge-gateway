package router

import (
	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/internal/config"
	"github.com/okuc/llm-bridge-gateway/internal/converter"
	"github.com/okuc/llm-bridge-gateway/internal/handler"
	"github.com/okuc/llm-bridge-gateway/internal/proxy"
)

// Router 路由管理器
type Router struct {
	config        *config.Config
	converter     *converter.ProtocolConverter
	proxy         *proxy.Proxy
	streamHandler *proxy.StreamHandler
}

// New 创建路由管理器
func New(config *config.Config, converter *converter.ProtocolConverter, proxyInstance *proxy.Proxy) *Router {
	return &Router{
		config:        config,
		converter:     converter,
		proxy:         proxyInstance,
		streamHandler: proxy.NewStreamHandler(converter),
	}
}

// RegisterRoutes 注册所有路由
func (r *Router) RegisterRoutes(engine *gin.Engine) {
	// 注册健康检查
	if r.config.Health.Enabled {
		engine.GET(r.config.Health.Path, handler.HealthCheck())
	}

	registeredPaths := make(map[string]bool)
	for _, route := range r.config.Routes {
		if !route.Enabled {
			continue
		}
		if registeredPaths[route.Input.Path] {
			continue
		}
		registeredPaths[route.Input.Path] = true

		// LLM API 固定使用 POST 方法
		engine.POST(route.Input.Path, r.createHandler(route))
	}
}

// createHandler 创建路由处理器
func (r *Router) createHandler(route config.RouteConfig) gin.HandlerFunc {
	h := handler.NewGatewayHandler(r.converter, r.proxy, &route, r.streamHandler)
	return h.HandleRequest
}

// MatchRoute 匹配路由
func (r *Router) MatchRoute(path string) *config.RouteConfig {
	for _, route := range r.config.Routes {
		if !route.Enabled {
			continue
		}

		if route.Input.Path == path {
			return &route
		}
	}

	return nil
}
