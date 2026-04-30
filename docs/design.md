# 大模型协议转换中间件 - 技术设计文档

## DES-001 系统架构

LLM Bridge Gateway 采用单进程 HTTP 网关架构：Gin Router 接收请求，Gateway Handler 编排协议转换与代理转发，OpenTrans SDK 执行协议转换，Proxy 使用 `net/http` 调用目标 API，Config Loader/Validator 负责启动时配置加载与校验。

```
Client -> Gin Router -> Gateway Handler -> OpenTrans Converter -> HTTP Proxy -> LLM API
                                  ^                              |
                                  |---- response conversion <----|
```

对应需求：REQ-F-001、REQ-F-002、REQ-F-006。

## DES-002 协议模型

内部协议标识统一为：

| 配置值         | 协议                        |
| ----------- | ------------------------- |
| `openai`    | OpenAI Chat Completions   |
| `responses` | OpenAI Responses          |
| `claude`    | Anthropic Claude Messages |
| `gemini`    | Google Gemini             |

`internal/converter` 将配置字符串映射为 OpenTrans 协议常量，并注册四种协议两两之间的转换器。请求方向是 `source -> target`，响应方向是 `target -> source`。

对应需求：REQ-F-001、REQ-F-002。

## DES-003 非流式请求流程

1. `cmd/gateway/main.go` 加载并验证配置，初始化 logger、converter、proxy 和 router。
2. `internal/router/router.go` 为每个启用路由注册一个 POST handler。
3. `internal/handler/handler.go` 限制请求体 10MB，读取 JSON body，识别 source/target 协议。
4. Handler 调用 `ConvertRequest` 将客户端协议转换为上游协议。
5. `internal/proxy/proxy.go` 转发到 `output.base_url`。
6. Handler 读取上游响应，2xx 响应调用 `ConvertResponse(target, source)`，非 2xx 响应透传状态和 body。

对应需求：REQ-F-002、REQ-F-006、REQ-F-007。

## DES-004 流式请求流程

当请求体 `stream` 为 `true` 且上游响应 `Content-Type` 包含 `text/event-stream` 时，`internal/proxy/stream.go` 使用 scanner 读取 SSE 行，对 `data:` 内容调用 `ConvertStreamEvent(target, source)`，逐事件 flush 给客户端，并保留 `data: [DONE]`。

对应需求：REQ-F-003。

## DES-005 配置设计

配置字段以实际实现为准：

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

当前版本不支持 `endpoint + path` 拆分，也不支持 `default_target`。所有启用路由的 `input.path` 必须唯一；如果需要同一种输入协议转发到不同目标，应配置不同路径。

对应需求：REQ-F-004、REQ-F-005、REQ-NF-003。

## DES-006 Provider 认证与代理

Proxy 层集中处理目标协议认证：

- Claude：`x-api-key` 和 `anthropic-version: 2023-06-01`
- OpenAI/Responses：`Authorization: Bearer <api_key>`
- Gemini：URL query 参数 `key=<api_key>`

当前仅透传 `X-Request-ID`。如需透传更多 Header，应在安全审查后明确白名单。

对应需求：REQ-F-006。

## DES-007 错误处理

统一错误响应结构：

```json
{
  "error": {
    "code": "CONVERSION_ERROR",
    "message": "Failed to convert request body",
    "details": "..."
  }
}
```

错误码包括 `INVALID_REQUEST`、`CONVERSION_ERROR`、`UPSTREAM_ERROR`、`TIMEOUT_ERROR`、`RATE_LIMIT_ERROR`、`NOT_FOUND`、`INTERNAL_ERROR`。当前 `RATE_LIMIT_ERROR` 仅保留常量，尚未实现限流。

对应需求：REQ-F-007。

## DES-008 健康检查

启用 `health.enabled` 后注册 `GET health.path`，默认返回服务名和 `ok` 状态。

对应需求：REQ-NF-001。

## DES-009 日志

`pkg/logger` 基于 `log/slog` 初始化全局 logger。`internal/middleware/logger.go` 记录请求方法、路径、状态码、耗时、客户端 IP、User-Agent、响应大小和 `X-Request-ID`。

对应需求：REQ-NF-002。

## DES-010 重试与超时

`internal/proxy/retry.go` 对转发错误和 5xx 响应按 `route.retry.max_attempts` 重试，等待时间为 `backoff * attempt`。Handler 使用 `context.WithTimeout` 应用 `route.timeout`。

对应需求：REQ-NF-003。

## DES-011 部署

构建产物是单个 Go 二进制。Dockerfile 使用多阶段构建，docker-compose 挂载 `config` 和 `logs` 并配置健康检查。

对应需求：REQ-NF-004。

## DES-012 性能设计目标

网关只在请求/响应边界做 JSON 协议转换和 HTTP 转发，不引入持久存储。性能目标需要后续 benchmark、并发压测和内存观测验证。

对应需求：REQ-NF-005。

## DES-013 Metrics 预留

配置结构保留 `metrics` 字段，但当前没有注册 `/metrics` handler，也没有引入 Prometheus client。该项属于后续扩展。

对应需求：REQ-NF-006。
