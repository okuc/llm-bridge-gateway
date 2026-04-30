# 大模型协议转换中间件 - 需求文档

## 1. 项目概述

LLM Bridge Gateway 是一个轻量级 Go HTTP 网关，用于在不同大模型 API 协议之间转换请求、响应和 SSE 流式事件，降低客户端迁移和多模型接入成本。

支持的协议范围统一为四种：

- OpenAI Chat Completions（`openai`）
- OpenAI Responses（`responses`）
- Anthropic Claude Messages（`claude`）
- Google Gemini（`gemini`）

## 2. 功能需求

### REQ-F-001 协议转换矩阵

网关应支持四种协议之间的有向转换，排除同协议直通后共 12 条转换路径：

| 输入协议 | 输出协议 |
| --- | --- |
| OpenAI Chat Completions | Anthropic Claude Messages |
| OpenAI Chat Completions | OpenAI Responses |
| OpenAI Chat Completions | Google Gemini |
| Anthropic Claude Messages | OpenAI Chat Completions |
| Anthropic Claude Messages | OpenAI Responses |
| Anthropic Claude Messages | Google Gemini |
| OpenAI Responses | OpenAI Chat Completions |
| OpenAI Responses | Anthropic Claude Messages |
| OpenAI Responses | Google Gemini |
| Google Gemini | OpenAI Chat Completions |
| Google Gemini | Anthropic Claude Messages |
| Google Gemini | OpenAI Responses |

### REQ-F-002 请求/响应双向转换

网关应将客户端请求从输入协议转换为目标 API 协议，并将目标 API 响应反向转换回客户端协议。核心字段包括 `model`、消息内容、`temperature`、token 限制、工具调用、停止条件和响应格式等；具体字段能力以 OpenTrans SDK 支持范围为准。

### REQ-F-003 SSE 流式响应转换

当请求体包含 `"stream": true` 且上游返回 `text/event-stream` 时，网关应逐个转换 SSE `data:` 事件，并保留 `[DONE]` 结束标记。

### REQ-F-004 YAML 路由配置

网关应从 YAML 配置加载服务、日志、健康检查和路由配置。当前版本按 `input.path` 精确匹配路由，所有 `enabled: true` 的路由必须使用唯一 `input.path`。

### REQ-F-005 API Key 环境变量注入

路由的 `output.api_key` 应支持 `${ENV_VAR}` 环境变量引用。启用路由的 API key 必须在启动校验时解析为非空真实值。

### REQ-F-006 上游代理转发

网关应将转换后的请求转发到 `output.base_url`，并按目标协议设置认证信息：Claude 使用 `x-api-key`，OpenAI/Responses 使用 `Authorization: Bearer`，Gemini 使用 `key` query 参数。

### REQ-F-007 错误响应

网关应返回统一 JSON 错误结构，至少覆盖 `INVALID_REQUEST`、`CONVERSION_ERROR`、`UPSTREAM_ERROR`、`TIMEOUT_ERROR`、`NOT_FOUND`、`INTERNAL_ERROR`。

## 3. 非功能需求

### REQ-NF-001 健康检查

网关应提供可配置的健康检查端点，默认 `/health`。

### REQ-NF-002 日志

网关应输出结构化请求日志，支持 `debug`、`info`、`warn`、`error` 级别和 `json`/`text` 格式。

### REQ-NF-003 重试与超时

每条路由应支持请求超时和重试配置。5xx 上游响应和转发错误可按配置重试；超时应返回 `TIMEOUT_ERROR`。

### REQ-NF-004 部署

网关应支持单二进制、Docker 和 Docker Compose 部署。

### REQ-NF-005 性能目标

目标是协议转换开销低于 50ms、支持至少 100 并发、常规运行内存低于 100MB。当前仓库尚未提供基准测试证明这些指标，需通过后续 benchmark 和压测验证。

### REQ-NF-006 可观测性扩展

当前版本提供健康检查和结构化日志。Prometheus `/metrics` 是预留配置项，尚未实现处理器。

## 4. 用户场景

### SCN-001 OpenAI 客户端访问 Claude

现有 OpenAI SDK 应用无需修改请求格式，通过网关路径访问 Claude 上游。

### SCN-002 Claude 客户端访问 OpenAI

使用 Claude Messages 格式的客户端可通过网关访问 OpenAI Chat Completions 上游。

### SCN-003 Chat Completions 迁移到 Responses

客户端可先通过独立路径将 OpenAI Chat Completions 请求转换为 OpenAI Responses 上游调用，逐步验证迁移。

### SCN-004 流式响应

客户端发送 `stream: true` 后，网关应转换上游 SSE 事件并持续返回给客户端。

## 5. 约束和假设

- 使用 Go 开发。
- 协议语义转换依赖 OpenTrans SDK。
- 客户端请求需符合对应源协议 JSON 格式。
- 目标 API 服务和网络链路可用。
- 当前路由模型是路径路由，不按模型名、Header 或请求体动态选择目标。

## 6. 追溯矩阵

| 需求 ID | 设计章节 | 计划任务 | 测试覆盖 | 用户手册 |
| --- | --- | --- | --- | --- |
| REQ-F-001 | DES-001, DES-002 | PLAN-002, PLAN-005 | `TestInit`, `TestAllProtocolConversions`（部分 SDK 样例） | UG-001, UG-002, UG-003 |
| REQ-F-002 | DES-002, DES-003 | PLAN-002 | `TestConvertRequest_*`, `TestConvertResponse_*`, `TestHTTPHandler_OpenAIToClaude` | UG-001, UG-002, UG-003 |
| REQ-F-003 | DES-004 | PLAN-003 | `TestConvertStreamEvent_*`, `TestIsStreamRequest` | UG-004 |
| REQ-F-004 | DES-005 | PLAN-001 | `TestValidatorValidate` | UG-005 |
| REQ-F-005 | DES-005 | PLAN-001 | `TestValidatorValidate/unresolved api key placeholder` | UG-005 |
| REQ-F-006 | DES-006 | PLAN-003 | `TestForwardProviderAuth` | UG-001, UG-002, UG-003 |
| REQ-F-007 | DES-007 | PLAN-003 | 待补错误响应集成测试 | UG-006 |
| REQ-NF-001 | DES-008 | PLAN-001 | 待补健康检查 handler 测试 | UG-006 |
| REQ-NF-002 | DES-009 | PLAN-001 | 待补日志中间件测试 | UG-007 |
| REQ-NF-003 | DES-010 | PLAN-003 | `TestValidatorValidate`, 待补重试集成测试 | UG-005 |
| REQ-NF-004 | DES-011 | PLAN-004 | Dockerfile/compose 静态覆盖，待补容器启动验证 | UG-008 |
| REQ-NF-005 | DES-012 | PLAN-006 | 待补 benchmark/压测 | UG-009 |
| REQ-NF-006 | DES-013 | PLAN-006 | 待实现 metrics 后补充 | UG-009 |
