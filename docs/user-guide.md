# LLM Bridge Gateway 使用手册

## 简介

LLM Bridge Gateway 是一个轻量级大模型协议转换网关（LLM gateway / AI API gateway / OpenAI compatible proxy），支持 OpenAI Chat Completions、OpenAI Responses、Anthropic Claude Messages、Google Gemini 四种协议之间的请求、响应和 SSE 流式事件转换。

它适合把 OpenAI SDK、Claude 客户端、Gemini API、Agent 工具、AI IDE、聊天机器人和内部服务接到不同模型提供商上。相比完整 LLMOps 平台，本项目聚焦协议转换中间件：轻量、可自托管、易二次开发。

当前版本按路径匹配路由：每个启用路由配置一个唯一的 `input.path`，网关不会按 model、Header 或请求体自动选择目标。

## 快速开始

### 1. 构建

```bash
go mod tidy
go build -o gateway cmd/gateway/main.go
```

或使用 Docker：

```bash
docker build -t llm-bridge-gateway .
```

### 2. 配置

```bash
cp config/config.example.yaml config/config.yaml
```

设置环境变量：

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="AIza..."
```

Windows 命令行可使用：

```bat
set ANTHROPIC_API_KEY=sk-ant-...
set OPENAI_API_KEY=sk-...
set GEMINI_API_KEY=AIza...
```

### 3. 启动

```bash
./gateway -config ./config/config.yaml
go run cmd/gateway/main.go -config ./config/config.yaml
```

健康检查：

```bash
curl http://localhost:18168/health
```

## 配置说明（UG-005）

```yaml
server:
  host: "0.0.0.0"
  port: 18168
  read_timeout: 30s
  write_timeout: 60s
  idle_timeout: 120s
logging:
  level: "info"
  format: "json"
  output: "stdout"
routes:
  - name: "openai-to-claude"
    enabled: true
    input:
      protocol: "openai"
      path: "/v1/chat/completions"
    output:
      protocol: "claude"
      base_url: "https://api.anthropic.com/v1/messages"
      api_key: "${ANTHROPIC_API_KEY}"
      model: "claude-sonnet-4-20250514"
    timeout: 60s
    retry:
      max_attempts: 3
      backoff: 1s
health:
  enabled: true
  path: "/health"
metrics:
  enabled: false
  path: "/metrics"
```

配置规则：

- `routes[].input.protocol` 可选 `openai`、`responses`、`claude`、`gemini`。
- `routes[].output.base_url` 是目标 API 的完整 URL。
- 所有 `enabled: true` 的 `routes[].input.path` 必须唯一。
- `routes[].output.api_key` 支持 `${ENV_VAR}`，启用路由启动时必须解析为真实值。
- `metrics` 当前为预留配置，尚未注册 `/metrics` 处理器。

## 使用场景

### UG-001 OpenAI 客户端访问 Claude

对应需求：REQ-F-001、REQ-F-002、REQ-F-006。

```yaml
routes:
  - name: "openai-to-claude"
    enabled: true
    input:
      protocol: "openai"
      path: "/v1/chat/completions"
    output:
      protocol: "claude"
      base_url: "https://api.anthropic.com/v1/messages"
      api_key: "${ANTHROPIC_API_KEY}"
      model: "claude-sonnet-4-20250514"
```

```python
import openai

client = openai.OpenAI(
    base_url="http://localhost:18168/v1",
    api_key="dummy"
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)
```

### UG-002 Claude 客户端访问 OpenAI

对应需求：REQ-F-001、REQ-F-002、REQ-F-006。

```yaml
routes:
  - name: "claude-to-openai"
    enabled: true
    input:
      protocol: "claude"
      path: "/v1/messages"
    output:
      protocol: "openai"
      base_url: "https://api.openai.com/v1/chat/completions"
      api_key: "${OPENAI_API_KEY}"
      model: "gpt-4o"
```

### UG-003 Chat Completions 迁移到 Responses

对应需求：REQ-F-001、REQ-F-002、REQ-F-006。

```yaml
routes:
  - name: "openai-to-responses"
    enabled: true
    input:
      protocol: "openai"
      path: "/v1/openai/responses"
    output:
      protocol: "responses"
      base_url: "https://api.openai.com/v1/responses"
      api_key: "${OPENAI_API_KEY}"
      model: "gpt-4o"
```

请求示例：

```bash
curl http://localhost:18168/v1/openai/responses \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello!"}]}'
```

### UG-004 流式响应

对应需求：REQ-F-003。

```bash
curl http://localhost:18168/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Write a story"}],
    "stream": true
  }'
```

响应为 SSE：

```text
data: {"choices":[{"delta":{"content":"Once"}}]}
data: {"choices":[{"delta":{"content":" upon"}}]}
data: [DONE]
```

## API 参考（UG-006）

| 端点               | 方法   | 说明     |
| ---------------- | ---- | ------ |
| 配置的 `input.path` | POST | 协议转换入口 |
| `/health`        | GET  | 健康检查   |

健康检查响应：

```json
{
  "status": "ok",
  "service": "llm-bridge-gateway"
}
```

错误响应：

```json
{
  "error": {
    "code": "CONVERSION_ERROR",
    "message": "Failed to convert request body",
    "details": "..."
  }
}
```

错误码：`INVALID_REQUEST`、`CONVERSION_ERROR`、`UPSTREAM_ERROR`、`TIMEOUT_ERROR`、`NOT_FOUND`、`INTERNAL_ERROR`。

## 日志与排障（UG-007）

开启调试日志：

```yaml
logging:
  level: "debug"
  format: "text"
  output: "stdout"
```

常见问题：

- 启动失败：检查 YAML、环境变量、启用路由路径是否重复。
- `CONVERSION_ERROR`：检查请求 JSON 是否符合源协议格式。
- `UPSTREAM_ERROR`：检查 `output.base_url`、API key 和网络。
- `TIMEOUT_ERROR`：增大路由 `timeout` 或检查上游响应耗时。

## 部署指南（UG-008）

直接运行：

```bash
go build -o gateway cmd/gateway/main.go
./gateway -config ./config/config.yaml
```

Docker：

```bash
docker build -t llm-bridge-gateway .
docker run -d \
  -p 18168:18168 \
  -v ./config:/app/config \
  -e ANTHROPIC_API_KEY=sk-ant-... \
  -e OPENAI_API_KEY=sk-... \
  -e GEMINI_API_KEY=AIza... \
  llm-bridge-gateway
```

Docker Compose：

```bash
docker compose up -d
docker compose logs -f
docker compose down
```

## 性能与监控（UG-009）

协议转换主要依赖 OpenTrans SDK，目标是低延迟、低内存开销。当前仓库尚未提供 benchmark 或压测报告；生产使用前应根据真实请求体和上游延迟进行压测。

当前版本支持 `/health` 和结构化日志；Prometheus `/metrics` 尚未实现。
