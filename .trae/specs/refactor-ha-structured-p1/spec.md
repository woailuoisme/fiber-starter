# 结构化高可用重构（P1）Spec

## Why
当前工程已完成 P0 稳定性基线，但仍存在全局状态与 DI 并存、request_id 与日志字段不一致、队列 worker 缺少独立运维入口等结构性风险，会降低可测性与 HA 部署可控性。

## What Changes
- 逐步去全局状态：减少业务层对 `config.GlobalConfig` 与 `database.DB` 的直接依赖，收敛到 DI 与显式依赖传递
- 统一 request_id：明确 request_id 的“来源、存储、回写、读取”策略，并统一 access log / error log / business log 字段
- 增加队列 worker 运维模型：新增 `queue:work` CLI，以独立进程方式运行 asynq server，并纳入统一优雅退出
- **BREAKING**：禁止在业务层新增对 `config.GlobalConfig` / `database.DB` 的直接引用（违反工程契约的代码需要调整）

## Impact
- Affected specs: 结构化工程化（可维护/可测试/可演进）、生产级 HA 基线（运维模型）
- Affected code:
  - 配置与启动：`config/config.go`、`bootstrap/app.go`
  - DB：`database/connection.go`、依赖 DB 的 service / command
  - HTTP 中间件：`app/http/middleware/setup.go`、`app/http/middleware/error_handler.go`
  - 队列：`app/services/queue.go`、`command/*`（新增 worker 命令）

## ADDED Requirements
### Requirement: 依赖注入优先
系统 SHALL 在业务层通过 DI 或显式参数传递获取 Config 与 DB 依赖，而不是直接读取 `config.GlobalConfig` 或 `database.DB`。

#### Scenario: Web 请求链路使用 DB
- **WHEN** handler/service 需要访问数据库
- **THEN** 其依赖由容器注入的 `*database.Connection` 或 `*sql.DB` 提供
- **AND** 业务层不出现 `config.GlobalConfig` 与 `database.DB` 的直接引用

#### Scenario: CLI 命令使用 DB 与配置
- **WHEN** 执行迁移/seed/worker 等命令
- **THEN** 命令通过容器构造所需依赖（Config、Connection、Logger 等）
- **AND** 不依赖全局单例可完成单元测试（允许使用 stub/mocks）

### Requirement: request_id 统一策略
系统 SHALL 对每个请求提供唯一 request_id，并保证该值在 access log、错误日志与业务日志中使用同一字段名与读取方式。

#### Scenario: 请求未携带 request_id
- **WHEN** 请求未携带 `X-Request-ID`
- **THEN** 系统生成新的 request_id
- **AND** 将其写入响应头 `X-Request-ID`
- **AND** access log 与错误日志包含字段 `request_id`

#### Scenario: 请求携带 request_id
- **WHEN** 请求携带 `X-Request-ID`
- **THEN** 系统沿用该 request_id
- **AND** 该值在 Fiber request context 中以统一 key 存储，供所有日志读取

### Requirement: 队列 worker 独立运行命令
系统 SHALL 提供 `queue:work` CLI 命令以独立进程方式启动 asynq worker，并在收到 SIGINT/SIGTERM 时优雅退出。

#### Scenario: 启动 worker
- **WHEN** 执行 `queue:work`
- **THEN** 进程启动 asynq server 并开始消费队列
- **AND** 使用与主服务一致的结构化日志输出

#### Scenario: 优雅退出 worker
- **WHEN** 进程收到 SIGINT/SIGTERM
- **THEN** worker 停止接收新任务并尝试完成在途任务（按 asynq Shutdown 语义）
- **AND** 释放相关资源并退出为 0

## MODIFIED Requirements
### Requirement: 日志字段一致性
系统 SHALL 保证 access log、错误日志、业务日志至少包含：`request_id`、`path`、`method`、`status`、`latency`（若上下文可得），并保持字段命名一致。

## REMOVED Requirements
### Requirement: 业务层可直接使用关键全局单例
**Reason**: 全局单例会降低可测性并引入多实例/多环境风险，不符合结构化与 HA 目标。
**Migration**: 通过 DI 注入 Config/DB/Logger；对未迁移的旧代码采用“适配层”过渡并逐步替换。
