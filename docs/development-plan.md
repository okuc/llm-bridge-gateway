# 大模型协议转换中间件 - 开发计划

## 1. 当前范围

LLM Bridge Gateway 当前版本目标是提供一个轻量级 Go HTTP 网关，支持 OpenAI Chat Completions、OpenAI Responses、Anthropic Claude Messages、Google Gemini 四种协议之间的路径路由、请求/响应转换、SSE 流式事件转换、重试、超时、健康检查、日志和 Docker 部署。

## 2. 计划任务

### PLAN-001 基础框架与配置

对应需求：REQ-F-004、REQ-F-005、REQ-NF-001、REQ-NF-002。

- Go module 和目录结构
- Viper 配置加载
- 环境变量展开
- 启用路由校验：协议、路径唯一性、URL、API key
- Gin server、健康检查、CORS、日志、Recovery

交付文件：`cmd/gateway/main.go`、`internal/config/`、`internal/middleware/`、`internal/handler/health.go`、`config/config.example.yaml`。

### PLAN-002 协议转换

对应需求：REQ-F-001、REQ-F-002。

- 集成 OpenTrans SDK
- 注册四协议两两转换
- 请求转换、响应反向转换、同协议直通
- 协议字符串映射和 stream 标识检测

交付文件：`internal/converter/converter.go`、`internal/converter/converter_test.go`。

### PLAN-003 代理、流式和错误处理

对应需求：REQ-F-003、REQ-F-006、REQ-F-007、REQ-NF-003。

- HTTP proxy 转发
- Provider 认证注入
- `X-Request-ID` 透传
- 路由级 timeout
- 5xx/转发错误重试
- 非流式响应转换
- SSE `data:` 事件转换
- 标准错误响应

交付文件：`internal/proxy/`、`internal/handler/handler.go`、`internal/handler/errors.go`。

### PLAN-004 部署与运行

对应需求：REQ-NF-004。

- Makefile 构建、测试、vet、coverage
- Dockerfile 多阶段构建
- docker-compose 服务和健康检查

交付文件：`Makefile`、`Dockerfile`、`docker-compose.yml`。

### PLAN-005 测试覆盖

对应需求：所有功能需求。

- 配置校验测试
- 协议转换测试
- Provider auth/proxy 测试
- Recovery 稳定性测试
- Handler 集成测试
- 后续补充：错误响应集成测试、健康检查测试、重试测试、完整 12 路径可用性测试

交付文件：`internal/**/*_test.go`、`test/integration/gateway_test.go`。

### PLAN-006 文档与追溯

对应需求：所有需求。

- 需求 ID 化
- 设计章节 ID 化
- 用户手册场景 ID 化
- 需求-设计-计划-测试-用户手册追溯矩阵
- README 快速上手同步
- CLAUDE.md 开发约束同步

交付文件：`docs/requirements.md`、`docs/design.md`、`docs/development-plan.md`、`docs/user-guide.md`、`README.md`、`CLAUDE.md`。

## 3. 已知限制

- 当前路由模型是 `input.path` 精确匹配，不按 model、Header 或请求体动态调度。
- `metrics` 是配置预留，尚未实现 `/metrics` handler。
- 性能目标尚缺 benchmark 和压测数据。
- OpenTrans SDK 对部分协议样例有严格格式要求，测试应使用 SDK 支持的真实协议形状。

## 4. 后续增强

- 实现 Prometheus metrics handler。
- 增加 benchmark 和并发压测。
- 增加可选 Header 白名单透传配置。
- 增加按 model 或 Header 的路由策略，但需避免破坏当前简单路径路由模型。
- 增加错误响应、重试、健康检查的集成测试。
