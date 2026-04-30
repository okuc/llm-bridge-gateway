package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Loader 配置加载器
type Loader struct {
	viper *viper.Viper
}

// NewLoader 创建配置加载器
func NewLoader() *Loader {
	return &Loader{
		viper: viper.New(),
	}
}

// Load 加载配置文件
func (l *Loader) Load(configPath string) (*Config, error) {
	// 设置配置文件路径
	if configPath != "" {
		l.viper.SetConfigFile(configPath)
	} else {
		l.viper.SetConfigName("config")
		l.viper.SetConfigType("yaml")
		l.viper.AddConfigPath("./config")
		l.viper.AddConfigPath(".")
	}

	// 读取环境变量
	l.viper.AutomaticEnv()
	l.viper.SetEnvPrefix("LLM_GATEWAY")
	l.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := l.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// 解析配置
	var config Config
	if err := l.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 处理环境变量引用
	l.resolveEnvVars(&config)

	// 设置默认值
	l.setDefaults(&config)

	return &config, nil
}

// resolveEnvVars 解析配置中的环境变量引用
func (l *Loader) resolveEnvVars(config *Config) {
	for i := range config.Routes {
		route := &config.Routes[i]
		route.Output.APIKey = l.expandEnv(route.Output.APIKey)
	}
}

// expandEnv 展开环境变量引用
func (l *Loader) expandEnv(value string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envKey := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
		if envValue := os.Getenv(envKey); envValue != "" {
			return envValue
		}
		// 环境变量缺失，记录警告
		fmt.Printf("WARNING: Environment variable %s is not set\n", envKey)
	}
	return value
}

// setDefaults 设置默认值
func (l *Loader) setDefaults(config *Config) {
	// 服务器默认值
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 18168
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 60 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 120 * time.Second
	}

	// 日志默认值
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	// 健康检查默认值
	if config.Health.Path == "" {
		config.Health.Path = "/health"
	}

	// 指标默认值
	if config.Metrics.Path == "" {
		config.Metrics.Path = "/metrics"
	}

	// 路由默认值
	for i := range config.Routes {
		route := &config.Routes[i]
		if route.Timeout == 0 {
			route.Timeout = 60 * time.Second
		}
		if route.Retry.MaxAttempts == 0 {
			route.Retry.MaxAttempts = 3
		}
		if route.Retry.Backoff == 0 {
			route.Retry.Backoff = 1 * time.Second
		}
	}
}
