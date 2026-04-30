package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Validator 配置验证器
type Validator struct{}

// NewValidator 创建配置验证器
func NewValidator() *Validator {
	return &Validator{}
}

// Validate 验证配置
func (v *Validator) Validate(config *Config) error {
	// 验证服务器配置
	if err := v.validateServer(&config.Server); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	// 验证日志配置
	if err := v.validateLogging(&config.Logging); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	// 验证路由配置
	if err := v.validateRoutes(config.Routes); err != nil {
		return fmt.Errorf("routes config: %w", err)
	}

	return nil
}

// validateServer 验证服务器配置
func (v *Validator) validateServer(config *ServerConfig) error {
	if config.Port < 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port: %d", config.Port)
	}
	return nil
}

// validateLogging 验证日志配置
func (v *Validator) validateLogging(config *LoggingConfig) error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[config.Level] {
		return fmt.Errorf("invalid log level: %s", config.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[config.Format] {
		return fmt.Errorf("invalid log format: %s", config.Format)
	}

	validOutputs := map[string]bool{
		"stdout": true,
		"file":   true,
	}
	if !validOutputs[config.Output] {
		return fmt.Errorf("invalid log output: %s", config.Output)
	}

	if config.Output == "file" && config.FilePath == "" {
		return fmt.Errorf("file_path is required when output is file")
	}

	return nil
}

// validateRoutes 验证路由配置
func (v *Validator) validateRoutes(routes []RouteConfig) error {
	if len(routes) == 0 {
		return fmt.Errorf("at least one route must be configured")
	}

	names := make(map[string]bool)
	paths := make(map[string]string)
	enabledCount := 0

	for _, route := range routes {
		if route.Name == "" {
			return fmt.Errorf("route name is required")
		}
		if names[route.Name] {
			return fmt.Errorf("duplicate route name: %s", route.Name)
		}
		names[route.Name] = true

		if !route.Enabled {
			continue
		}
		enabledCount++

		if err := v.validateRoute(&route); err != nil {
			return fmt.Errorf("route %s: %w", route.Name, err)
		}

		if existing, exists := paths[route.Input.Path]; exists {
			return fmt.Errorf("duplicate enabled input path %s in routes %s and %s", route.Input.Path, existing, route.Name)
		}
		paths[route.Input.Path] = route.Name
	}

	if enabledCount == 0 {
		return fmt.Errorf("at least one route must be enabled")
	}

	return nil
}

// validateRoute 验证单个启用路由配置
func (v *Validator) validateRoute(route *RouteConfig) error {
	validProtocols := map[string]bool{
		ProtocolOpenAI:    true,
		ProtocolClaude:    true,
		ProtocolResponses: true,
		ProtocolGemini:    true,
	}

	if !validProtocols[route.Input.Protocol] {
		return fmt.Errorf("invalid input protocol: %s", route.Input.Protocol)
	}

	if !validProtocols[route.Output.Protocol] {
		return fmt.Errorf("invalid output protocol: %s", route.Output.Protocol)
	}

	if route.Input.Path == "" {
		return fmt.Errorf("input path is required")
	}
	if !strings.HasPrefix(route.Input.Path, "/") {
		return fmt.Errorf("input path must start with /: %s", route.Input.Path)
	}

	if route.Output.BaseURL == "" {
		return fmt.Errorf("output base_url is required")
	}

	parsedURL, err := url.Parse(route.Output.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid output base_url: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("output base_url must use http or https scheme")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("output base_url must include host")
	}

	if route.Output.APIKey == "" {
		return fmt.Errorf("output api_key is required")
	}
	if strings.HasPrefix(route.Output.APIKey, "${") && strings.HasSuffix(route.Output.APIKey, "}") {
		return fmt.Errorf("output api_key environment variable is not resolved: %s", route.Output.APIKey)
	}

	return nil
}
