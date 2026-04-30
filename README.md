# LLM Bridge Gateway

[中文](#中文) | [English](#english)

> Lightweight LLM protocol conversion gateway for OpenAI-compatible, Claude-compatible, Responses API and Gemini-compatible traffic.
> 
> 轻量级大模型协议转换网关：让 OpenAI、Claude、Responses API、Gemini 客户端和服务端互相兼容。

**Keywords:** LLM gateway, AI gateway, OpenAI compatible proxy, Claude API proxy, Anthropic adapter, Gemini API gateway, OpenAI Responses API, LLM protocol converter, model gateway, API gateway, self-hosted AI gateway, OpenAI to Claude, Claude to OpenAI, Gemini to OpenAI, SSE streaming proxy, 大模型网关, OpenAI 兼容代理, Claude 转 OpenAI, Gemini 协议转换, AI API 网关, 大模型协议转换中间件。

---

## 中文

### 为什么需要它？

很多工具只支持一种 API 格式：有的只会调用 OpenAI Chat Completions，有的需要 Anthropic Messages，有的开始迁移到 OpenAI Responses API，还有的要接 Google Gemini。**LLM Bridge Gateway** 提供一个可自托管的协议转换层，让现有客户端尽量少改代码即可访问不同模型服务。

相比更重的 LLMOps/计费/管理平台，本项目聚焦一件事：**协议转换 + 轻量代理 + 流式转发**。

### 核心特性

- **四协议互转**：OpenAI Chat Completions、OpenAI Responses、Anthropic Claude Messages、Google Gemini
- **OpenAI-compatible / Claude-compatible / Gemini-compatible**：适合现有 SDK、CLI、Agent 工具接入
- **双向转换**：请求 `source -> target`，响应 `target -> source`
- **流式响应转换**：支持 SSE `data:` 事件转换与 `[DONE]` 保留
- **路径路由**：通过 YAML 配置不同入口路径到不同上游模型 API
- **Provider 认证适配**：OpenAI Bearer Token、Claude `x-api-key`、Gemini `key` query
- **生产基础能力**：超时、5xx 重试、健康检查、结构化日志、优雅关闭
- **易部署**：单二进制、Docker、Docker Compose

### 适用场景

- 用 OpenAI SDK 调 Claude、Gemini 或 Responses API
- 用 Claude Messages 格式客户端访问 OpenAI-compatible API
- 在 OpenAI Chat Completions 和 OpenAI Responses API 之间渐进迁移
- 给 Agent、AI IDE、聊天机器人、内部工具提供统一协议入口
- 自托管一个轻量 AI API gateway / LLM gateway，而不是引入完整管理平台

### 支持的转换方向

| From                    | To                               |
| ----------------------- | -------------------------------- |
| OpenAI Chat Completions | Claude / Responses / Gemini      |
| Claude Messages         | OpenAI Chat / Responses / Gemini |
| OpenAI Responses        | OpenAI Chat / Claude / Gemini    |
| Gemini                  | OpenAI Chat / Claude / Responses |

协议转换能力以 [OpenTrans](https://github.com/xy200303/OpenTrans) SDK 支持范围为准。

### 快速开始

```bash
go mod tidy
cp config/config.example.yaml config/config.yaml
```

设置 API Key：

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="AIza..."
```

运行：

```bash
go run cmd/gateway/main.go -config ./config/config.yaml
```

健康检查：

```bash
curl http://localhost:18168/health
```

### 配置示例

当前版本按 `input.path` 精确匹配路由，所有启用路由的路径必须唯一。

```yaml
server:
  host: "0.0.0.0"
  port: 18168
  read_timeout: 30s
  write_timeout: 60s
  idle_timeout: 120s

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
```

协议值：`openai`、`responses`、`claude`、`gemini`。

### 调用示例

OpenAI 格式请求转 Claude：

```bash
curl http://localhost:18168/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

OpenAI Chat 转 Responses：

```bash
curl http://localhost:18168/v1/openai/responses \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello!"}]}'
```

流式响应：

```bash
curl http://localhost:18168/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Write a story"}],
    "stream": true
  }'
```

### 构建与测试

```bash
make build          # 构建 ./build/llm-bridge-gateway
make dev            # go run cmd/gateway/main.go -config ./config/config.yaml
make test           # go test -v ./...
make test-coverage  # 生成 coverage.html
make vet            # go vet ./...
make tidy           # go mod tidy
```

没有 `make` 时可直接运行等价命令：

```bash
go build -o ./build/llm-bridge-gateway ./cmd/gateway/main.go
go test ./...
go vet ./...
```

### Docker

```bash
docker build -t llm-bridge-gateway .
docker compose up --build
```

### 面向谁

- **最终用户**：想把现有 OpenAI/Claude/Gemini 客户端快速接到不同模型服务，不想大改业务代码。
- **开发者**：想要一个清晰可读、容易二次开发的 Go 协议转换网关，可按需要扩展路由、认证、观测、限流或更多协议。
- **团队/私有部署**：想自托管一个轻量 AI API gateway，作为内部工具、Agent、AI IDE、聊天机器人或服务端应用的统一入口。

本项目既可以直接作为独立网关使用，也可以作为 sidecar、中间件或协议适配层集成到现有系统中。

### 文档

- 需求与追溯矩阵：`docs/requirements.md`
- 技术设计：`docs/design.md`
- 开发计划：`docs/development-plan.md`
- 使用手册：`docs/user-guide.md`

### 当前限制

- 路由按路径匹配，不按 model、Header 或请求体动态调度。
- `metrics` 配置为预留项，当前尚未实现 `/metrics` handler。
- 性能目标尚未通过仓库内 benchmark/压测证明。

---

## English

### Why this project?

Many AI clients and developer tools are tied to one API shape: OpenAI Chat Completions, Anthropic Claude Messages, the newer OpenAI Responses API, or Google Gemini. **LLM Bridge Gateway** is a self-hosted compatibility layer that converts requests, responses and streaming events between these protocols.

Instead of being a full LLMOps platform, this project focuses on a small and practical core: **protocol conversion, lightweight proxying, and streaming passthrough**.

### Features

- **Four protocol families**: OpenAI Chat Completions, OpenAI Responses, Anthropic Claude Messages, Google Gemini
- **OpenAI-compatible, Claude-compatible, Gemini-compatible** gateway endpoints
- **Bidirectional conversion**: request `source -> target`, response `target -> source`
- **SSE streaming conversion**: converts `data:` events and preserves `[DONE]`
- **YAML route configuration**: map different inbound paths to different upstream APIs
- **Provider auth adaptation**: OpenAI Bearer token, Claude `x-api-key`, Gemini `key` query param
- **Production basics**: timeouts, 5xx retries, health check, structured logs, graceful shutdown
- **Easy deployment**: single binary, Docker, Docker Compose

### Use cases

- Call Claude, Gemini or Responses API from an OpenAI SDK client
- Connect Claude Messages clients to OpenAI-compatible backends
- Gradually migrate from OpenAI Chat Completions to the Responses API
- Provide a unified protocol entrypoint for agents, AI IDEs, chatbots and internal tools
- Run a lightweight self-hosted AI API gateway without adopting a full management platform

### Supported conversion directions

| From                    | To                               |
| ----------------------- | -------------------------------- |
| OpenAI Chat Completions | Claude / Responses / Gemini      |
| Claude Messages         | OpenAI Chat / Responses / Gemini |
| OpenAI Responses        | OpenAI Chat / Claude / Gemini    |
| Gemini                  | OpenAI Chat / Claude / Responses |

Conversion semantics depend on the capabilities of the [OpenTrans](https://github.com/xy200303/OpenTrans) SDK.

### Quick start

```bash
go mod tidy
cp config/config.example.yaml config/config.yaml
```

Set API keys:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GEMINI_API_KEY="AIza..."
```

Run:

```bash
go run cmd/gateway/main.go -config ./config/config.yaml
```

Health check:

```bash
curl http://localhost:18168/health
```

### Route configuration

Routes are matched by exact `input.path`. Every enabled route must have a unique path.

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
    timeout: 60s
    retry:
      max_attempts: 3
      backoff: 1s
```

Protocol values: `openai`, `responses`, `claude`, `gemini`.

### Examples

OpenAI-format request to Claude:

```bash
curl http://localhost:18168/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

OpenAI Chat to Responses API:

```bash
curl http://localhost:18168/v1/openai/responses \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello!"}]}'
```

Streaming:

```bash
curl http://localhost:18168/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Write a story"}],
    "stream": true
  }'
```

### Build and test

```bash
make build
make test
make vet
```

Without `make`:

```bash
go build -o ./build/llm-bridge-gateway ./cmd/gateway/main.go
go test ./...
go vet ./...
```

### Docker

```bash
docker build -t llm-bridge-gateway .
docker compose up --build
```

### Who is it for?

- **End users** who want existing OpenAI, Claude or Gemini clients to work with different model providers without rewriting application code.
- **Developers** who want a readable Go protocol conversion gateway that is easy to customize for routing, authentication, observability, rate limiting or additional protocols.
- **Teams and private deployments** that need a lightweight self-hosted AI API gateway for internal tools, agents, AI IDEs, chatbots or backend services.

It can be used directly as a standalone gateway, or embedded as a sidecar, middleware, or protocol adapter in a larger system.

### Docs

- Requirements and traceability matrix: `docs/requirements.md`
- Technical design: `docs/design.md`
- Development plan: `docs/development-plan.md`
- User guide: `docs/user-guide.md`

### Current limitations

- Routes are path-based; there is no model/header/body-based router yet.
- `metrics` is reserved in config, but `/metrics` is not implemented yet.
- Performance targets still need repository benchmarks and load tests.

## License

MIT License
