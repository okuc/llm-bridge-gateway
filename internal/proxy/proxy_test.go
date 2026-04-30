package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/okuc/llm-bridge-gateway/internal/config"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
)

func TestForwardProviderAuth(t *testing.T) {
	logger.Init("error", "text", "stdout", "")
	tests := []struct {
		name     string
		protocol string
		check    func(*testing.T, *http.Request)
	}{
		{
			name:     "claude auth headers",
			protocol: config.ProtocolClaude,
			check: func(t *testing.T, r *http.Request) {
				if got := r.Header.Get("x-api-key"); got != "test-key" {
					t.Fatalf("x-api-key = %q, want test-key", got)
				}
				if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
					t.Fatalf("anthropic-version = %q", got)
				}
			},
		},
		{
			name:     "openai authorization header",
			protocol: config.ProtocolOpenAI,
			check: func(t *testing.T, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("Authorization = %q, want Bearer test-key", got)
				}
			},
		},
		{
			name:     "responses authorization header",
			protocol: config.ProtocolResponses,
			check: func(t *testing.T, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("Authorization = %q, want Bearer test-key", got)
				}
			},
		},
		{
			name:     "gemini key query",
			protocol: config.ProtocolGemini,
			check: func(t *testing.T, r *http.Request) {
				if got := r.URL.Query().Get("key"); got != "test-key" {
					t.Fatalf("key query = %q, want test-key", got)
				}
				if r.URL.RawQuery == "&key=test-key" || r.URL.RawQuery == "?key=test-key" {
					t.Fatalf("malformed raw query: %q", r.URL.RawQuery)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.check(t, r)
				if got := r.Header.Get("X-Request-ID"); got != "req-123" {
					t.Fatalf("X-Request-ID = %q, want req-123", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
			}))
			defer server.Close()

			route := &config.RouteConfig{
				Name:    "test",
				Enabled: true,
				Input: config.InputConfig{
					Protocol: config.ProtocolOpenAI,
					Path:     "/v1/chat/completions",
				},
				Output: config.OutputConfig{
					Protocol: tt.protocol,
					BaseURL:  server.URL + "/target?existing=1",
					APIKey:   "test-key",
				},
				Timeout: time.Second,
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
			req.Header.Set("X-Request-ID", "req-123")

			resp, err := New().Forward(req, route, []byte(`{"model":"test"}`))
			if err != nil {
				t.Fatalf("Forward() error = %v", err)
			}
			defer resp.Body.Close()
		})
	}
}
