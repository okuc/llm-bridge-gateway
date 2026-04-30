package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	opentrans "github.com/xy200303/OpenTrans"
)

// ProtocolConverter 协议转换器
type ProtocolConverter struct{}

// New 创建协议转换器
func New() *ProtocolConverter {
	return &ProtocolConverter{}
}

// Init 初始化转换器
func (pc *ProtocolConverter) Init() {}

// ConvertRequest 转换请求体
// 支持双向转换：上行（客户端→目标API）和下行（目标API→客户端）
func (pc *ProtocolConverter) ConvertRequest(body []byte, source, target opentrans.Protocol) ([]byte, error) {
	if source == target {
		return body, nil
	}

	result, err := opentrans.ConvertRequestBody(body, source, target)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request from %s to %s: %w", source, target, err)
	}

	return result, nil
}

// ConvertResponse 转换响应体
// 支持双向转换：上行（目标API响应→客户端格式）和下行（客户端→目标API格式）
func (pc *ProtocolConverter) ConvertResponse(body []byte, source, target opentrans.Protocol) ([]byte, error) {
	if source == target {
		return body, nil
	}

	result, err := opentrans.ConvertResponseBody(body, source, target)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response from %s to %s: %w", source, target, err)
	}

	return result, nil
}

// ConvertStreamEvent 转换流式事件
func (pc *ProtocolConverter) ConvertStreamEvent(event []byte, source, target opentrans.Protocol) ([]byte, error) {
	if source == target {
		return event, nil
	}

	result, err := opentrans.ConvertStreamEventBody(event, source, target)
	if err != nil {
		return nil, fmt.Errorf("failed to convert stream event from %s to %s: %w", source, target, err)
	}

	return result, nil
}

// GetProtocol 从字符串获取协议类型
func GetProtocol(protocol string) (opentrans.Protocol, error) {
	switch strings.ToLower(protocol) {
	case "openai", "chat-completions":
		return opentrans.ProtocolOpenAI, nil
	case "claude", "anthropic":
		return opentrans.ProtocolClaude, nil
	case "responses":
		return opentrans.ProtocolResponses, nil
	case "gemini":
		return opentrans.ProtocolGemini, nil
	default:
		return "", fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// IsStreamRequest 检查是否是流式请求
func IsStreamRequest(body []byte) bool {
	var request struct {
		Stream           bool `json:"stream"`
		GenerationConfig struct {
			Stream bool `json:"stream"`
		} `json:"generationConfig"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return false
	}
	return request.Stream || request.GenerationConfig.Stream
}
