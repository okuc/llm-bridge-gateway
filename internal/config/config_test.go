package config

import (
	"strings"
	"testing"
	"time"
)

func validConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 18168,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Routes: []RouteConfig{
			validRoute("openai-to-claude", "/v1/chat/completions"),
		},
	}
}

func validRoute(name, path string) RouteConfig {
	return RouteConfig{
		Name:    name,
		Enabled: true,
		Input: InputConfig{
			Protocol: ProtocolOpenAI,
			Path:     path,
		},
		Output: OutputConfig{
			Protocol: ProtocolClaude,
			BaseURL:  "https://api.anthropic.com/v1/messages",
			APIKey:   "test-key",
		},
		Timeout: 60 * time.Second,
		Retry: RetryConfig{
			MaxAttempts: 3,
			Backoff:     time.Second,
		},
	}
}

func TestConfig_GetAddress(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "default address",
			config: Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 18168,
				},
			},
			expected: "0.0.0.0:18168",
		},
		{
			name: "custom address",
			config: Config{
				Server: ServerConfig{
					Host: "127.0.0.1",
					Port: 9090,
				},
			},
			expected: "127.0.0.1:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetAddress()
			if result != tt.expected {
				t.Errorf("GetAddress() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidatorValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "valid config",
		},
		{
			name: "invalid log level",
			mutate: func(cfg *Config) {
				cfg.Logging.Level = "trace"
			},
			wantErr: "invalid log level",
		},
		{
			name: "invalid log format",
			mutate: func(cfg *Config) {
				cfg.Logging.Format = "xml"
			},
			wantErr: "invalid log format",
		},
		{
			name: "invalid log output",
			mutate: func(cfg *Config) {
				cfg.Logging.Output = "stderr"
			},
			wantErr: "invalid log output",
		},
		{
			name: "file log output requires path",
			mutate: func(cfg *Config) {
				cfg.Logging.Output = "file"
				cfg.Logging.FilePath = ""
			},
			wantErr: "file_path is required",
		},
		{
			name: "invalid server port",
			mutate: func(cfg *Config) {
				cfg.Server.Port = 70000
			},
			wantErr: "invalid port",
		},
		{
			name: "duplicate enabled input path",
			mutate: func(cfg *Config) {
				cfg.Routes = append(cfg.Routes, validRoute("openai-to-gemini", "/v1/chat/completions"))
			},
			wantErr: "duplicate enabled input path",
		},
		{
			name: "disabled incomplete route is ignored",
			mutate: func(cfg *Config) {
				cfg.Routes = append(cfg.Routes, RouteConfig{Name: "disabled", Enabled: false})
			},
		},
		{
			name: "unresolved api key placeholder",
			mutate: func(cfg *Config) {
				cfg.Routes[0].Output.APIKey = "${ANTHROPIC_API_KEY}"
			},
			wantErr: "environment variable is not resolved",
		},
		{
			name: "invalid output url",
			mutate: func(cfg *Config) {
				cfg.Routes[0].Output.BaseURL = "api.anthropic.com/v1/messages"
			},
			wantErr: "must use http or https",
		},
		{
			name: "missing output host",
			mutate: func(cfg *Config) {
				cfg.Routes[0].Output.BaseURL = "https:///v1/messages"
			},
			wantErr: "must include host",
		},
		{
			name: "invalid input protocol",
			mutate: func(cfg *Config) {
				cfg.Routes[0].Input.Protocol = "invalid"
			},
			wantErr: "invalid input protocol",
		},
		{
			name: "input path must start with slash",
			mutate: func(cfg *Config) {
				cfg.Routes[0].Input.Path = "v1/chat/completions"
			},
			wantErr: "input path must start with /",
		},
		{
			name: "no enabled routes",
			mutate: func(cfg *Config) {
				cfg.Routes[0].Enabled = false
			},
			wantErr: "at least one route must be enabled",
		},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			if tt.mutate != nil {
				tt.mutate(cfg)
			}
			err := validator.Validate(cfg)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() unexpected error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() expected error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestProtocolConstants(t *testing.T) {
	if ProtocolOpenAI != "openai" {
		t.Errorf("Expected ProtocolOpenAI = 'openai', got %v", ProtocolOpenAI)
	}
	if ProtocolClaude != "claude" {
		t.Errorf("Expected ProtocolClaude = 'claude', got %v", ProtocolClaude)
	}
	if ProtocolResponses != "responses" {
		t.Errorf("Expected ProtocolResponses = 'responses', got %v", ProtocolResponses)
	}
	if ProtocolGemini != "gemini" {
		t.Errorf("Expected ProtocolGemini = 'gemini', got %v", ProtocolGemini)
	}
}
