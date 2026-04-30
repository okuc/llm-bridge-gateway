package config

import (
	"fmt"
	"time"
)

// Config 应用配置结构
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Logging LoggingConfig `mapstructure:"logging"`
	Routes  []RouteConfig `mapstructure:"routes"`
	Health  HealthConfig  `mapstructure:"health"`
	Metrics MetricsConfig `mapstructure:"metrics"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Name    string        `mapstructure:"name"`
	Enabled bool          `mapstructure:"enabled"`
	Input   InputConfig   `mapstructure:"input"`
	Output  OutputConfig  `mapstructure:"output"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retry   RetryConfig   `mapstructure:"retry"`
}

// InputConfig 输入配置
type InputConfig struct {
	Protocol string `mapstructure:"protocol"`
	Path     string `mapstructure:"path"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Protocol  string `mapstructure:"protocol"`
	BaseURL   string `mapstructure:"base_url"`
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int           `mapstructure:"max_attempts"`
	Backoff     time.Duration `mapstructure:"backoff"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// Protocol 协议类型常量
const (
	ProtocolOpenAI    = "openai"
	ProtocolClaude    = "claude"
	ProtocolResponses = "responses"
	ProtocolGemini    = "gemini"
)

// GetAddress 获取服务器监听地址
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
