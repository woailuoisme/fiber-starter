# 使用 Fiber v3 官方 Middleware 替换自定义实现 Spec

## Why
当前工程包含部分自定义中间件（request id、访问日志等），与 Fiber v3 官方 middleware 的能力重复，增加维护与一致性成本。统一改为官方 middleware，并补齐常用 middleware 组合，降低后续升级与运维风险。

## What Changes
- 替换 request id：移除自定义 `RequestID()` 中间件，改用 Fiber v3 官方 `requestid` middleware 生成/透传 `X-Request-ID`
- 替换访问日志：移除 setup.go 中的内联 zap access log 中间件，改用 Fiber v3 官方 `logger` middleware 输出访问日志
- 规范 middleware 顺序：将“抑制噪声/安全/限流/压缩/缓存相关/超时/日志”等按照推荐顺序挂载，避免重复或冲突
- 保留项目特有逻辑：与业务强相关的 JWT、参数校验、统一错误响应格式化等仍保留现有实现（不强行用官方 middleware 替代）

## Impact
- Affected specs: HTTP 中间件链路、日志/可观测性、请求链路追踪（request id）
- Affected code:
  - `app/http/middleware/setup.go`
  - `app/http/middleware/request_id.go`（预计删除或改为仅保留 request id 读取工具函数，视实现而定）
  - `tests/request_id_logging_test.go`（访问日志输出方式变化后需要调整）

## ADDED Requirements
### Requirement: 官方 RequestID
系统 SHALL 使用 Fiber v3 官方 RequestID middleware 为每个请求生成或透传请求标识，并回写响应头 `X-Request-ID`。

#### Scenario: 请求未携带 Request ID
- **WHEN** 客户端请求未携带 `X-Request-ID`
- **THEN** 服务端生成新的 request id
- **AND** 响应头包含 `X-Request-ID`

#### Scenario: 请求携带 Request ID
- **WHEN** 客户端请求携带 `X-Request-ID`
- **THEN** 服务端沿用该值
- **AND** 响应头的 `X-Request-ID` 与请求一致

### Requirement: 官方 Logger
系统 SHALL 使用 Fiber v3 官方 Logger middleware 输出访问日志，且日志内容包含：`method`、`path`、`status`、`latency`、`request_id`（来自 `X-Request-ID`）。

#### Scenario: 访问日志包含 Request ID
- **WHEN** 任意请求完成（成功或失败）
- **THEN** 访问日志中包含该请求的 request id（与响应头一致）

### Requirement: 常用 Middleware 组合
系统 SHALL 在 HTTP 链路中启用常用 middleware（至少包括）：favicon、recover、cors、helmet、limiter、compress、etag、timeout、requestid、logger。

## MODIFIED Requirements
### Requirement: 日志形态（访问日志）
访问日志从“zap 结构化字段”调整为“Fiber 官方 logger 输出格式化日志行”。

## REMOVED Requirements
### Requirement: 自定义 RequestID 中间件
**Reason**: 与 Fiber v3 官方 requestid middleware 能力重复，维护成本高。
**Migration**: 使用官方 requestid 替代，保留统一的 request id 读取工具函数（若仍被 error handler/其他日志使用）。

