package converter

import (
	"testing"

	opentrans "github.com/xy200303/OpenTrans"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestInit(t *testing.T) {
	c := New()
	c.Init()
}

func TestGetProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected opentrans.Protocol
		wantErr  bool
	}{
		{
			name:     "openai",
			input:    "openai",
			expected: opentrans.ProtocolOpenAI,
			wantErr:  false,
		},
		{
			name:     "chat-completions",
			input:    "chat-completions",
			expected: opentrans.ProtocolOpenAI,
			wantErr:  false,
		},
		{
			name:     "claude",
			input:    "claude",
			expected: opentrans.ProtocolClaude,
			wantErr:  false,
		},
		{
			name:     "anthropic",
			input:    "anthropic",
			expected: opentrans.ProtocolClaude,
			wantErr:  false,
		},
		{
			name:     "responses",
			input:    "responses",
			expected: opentrans.ProtocolResponses,
			wantErr:  false,
		},
		{
			name:     "gemini",
			input:    "gemini",
			expected: opentrans.ProtocolGemini,
			wantErr:  false,
		},
		{
			name:     "case insensitive",
			input:    "OpenAI",
			expected: opentrans.ProtocolOpenAI,
			wantErr:  false,
		},
		{
			name:     "invalid protocol",
			input:    "invalid",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetProtocol(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProtocol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("GetProtocol() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsStreamRequest(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{
			name:     "stream true",
			body:     `{"model": "gpt-4", "stream": true}`,
			expected: true,
		},
		{
			name:     "stream false",
			body:     `{"model": "gpt-4", "stream": false}`,
			expected: false,
		},
		{
			name:     "gemini generation config stream true",
			body:     `{"contents":[{"parts":[{"text":"hello"}]}],"generationConfig":{"stream":true}}`,
			expected: true,
		},
		{
			name:     "gemini generation config stream false",
			body:     `{"contents":[{"parts":[{"text":"hello"}]}],"generationConfig":{"stream":false}}`,
			expected: false,
		},
		{
			name:     "no stream field",
			body:     `{"model": "gpt-4"}`,
			expected: false,
		},
		{
			name:     "stream true no space",
			body:     `{"model":"gpt-4","stream":true}`,
			expected: true,
		},
		{
			name:     "invalid json",
			body:     `{`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStreamRequest([]byte(tt.body))
			if result != tt.expected {
				t.Errorf("IsStreamRequest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertRequest_SameProtocol(t *testing.T) {
	c := New()
	c.Init()

	body := []byte(`{"model": "gpt-4", "messages": [{"role": "user", "content": "hello"}]}`)

	result, err := c.ConvertRequest(body, opentrans.ProtocolOpenAI, opentrans.ProtocolOpenAI)
	if err != nil {
		t.Fatalf("ConvertRequest() error = %v", err)
	}

	if string(result) != string(body) {
		t.Errorf("ConvertRequest() = %v, want %v", string(result), string(body))
	}
}

func TestConvertResponse_SameProtocol(t *testing.T) {
	c := New()
	c.Init()

	body := []byte(`{"id": "chatcmpl-123", "choices": [{"message": {"content": "hello"}}]}`)

	result, err := c.ConvertResponse(body, opentrans.ProtocolOpenAI, opentrans.ProtocolOpenAI)
	if err != nil {
		t.Fatalf("ConvertResponse() error = %v", err)
	}

	if string(result) != string(body) {
		t.Errorf("ConvertResponse() = %v, want %v", string(result), string(body))
	}
}

func TestConvertStreamEvent_SameProtocol(t *testing.T) {
	c := New()
	c.Init()

	event := []byte(`{"choices": [{"delta": {"content": "hello"}}]}`)

	result, err := c.ConvertStreamEvent(event, opentrans.ProtocolOpenAI, opentrans.ProtocolOpenAI)
	if err != nil {
		t.Fatalf("ConvertStreamEvent() error = %v", err)
	}

	if string(result) != string(event) {
		t.Errorf("ConvertStreamEvent() = %v, want %v", string(result), string(event))
	}
}
