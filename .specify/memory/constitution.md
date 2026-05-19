<!--
Sync Impact Report
Version change: template placeholder -> 1.0.0
Modified principles:
- PRINCIPLE_1_NAME -> I. 代码质量与简洁性
- PRINCIPLE_2_NAME -> II. 测试标准与可验证性
- PRINCIPLE_3_NAME -> III. 用户体验与 API 一致性
- PRINCIPLE_4_NAME -> IV. 性能与高可用要求
- PRINCIPLE_5_NAME -> 合并到安全与运维约束
Added sections:
- 安全与运维约束
- 开发工作流与质量门禁
Removed sections:
- 模板占位章节 SECTION_2_NAME
- 模板占位章节 SECTION_3_NAME
Templates requiring updates:
- updated: .specify/templates/plan-template.md
- updated: .specify/templates/spec-template.md
- updated: .specify/templates/tasks-template.md
- n/a: .specify/templates/commands/*.md not present in this checkout
Runtime guidance updates:
- updated: README.md
Follow-up TODOs: none
-->
# Fiber Template Constitution

## Core Principles

### I. 代码质量与简洁性

生产代码 MUST 保持清晰、可维护、符合 Go 习惯用法。实现必须优先复用
仓库现有结构与 helper，避免为了兼容或抽象而新增无实际价值的层。

- Go 代码 MUST 通过 `gofmt` 或 `gofumpt` 格式化。
- 返回的错误 MUST 被检查；向上传播时 MUST 使用 `%w` 包裹上下文。
- 公共 API 返回的集合 MUST 使用具体空切片，禁止把空集合序列化为 `null`。
- 单个文件和函数 SHOULD 聚焦单一职责；重复逻辑必须先评估是否已有本地模式。
- 新增导出符号、全局状态、兼容层或 provider MUST 有明确使用场景。

理由：项目是长期演进的后端模板，代码质量比短期堆功能更重要。简洁实现能降低
维护成本、误用概率和后续重构风险。

### II. 测试标准与可验证性

任何影响外部行为、数据处理、配置、安全、性能或 provider 生命周期的变更
MUST 有对应测试或明确的验证记录。

- 测试文件 MUST 放在 `tests/` 目录，不得散落在 `internal/`、`configs/` 等生产目录。
- 新增或修改测试 MUST 使用 `github.com/stretchr/testify` 的 `assert` 或 `require`。
- API 契约、错误响应、路由行为、provider 降级和安全边界 MUST 优先用合同或集成测试保护。
- 修复 bug 时 MUST 先补能复现问题的测试，除非问题只能通过人工运行环境验证。
- 验证命令 MUST 使用 `rtk` 前缀，并在交付说明中区分已运行、未运行和无法运行的检查。

理由：该项目包含 HTTP、CLI、provider、缓存、队列、搜索、存储和文档生成等多个边界。
只有可复现测试和明确验证记录才能防止模板级回归。

### III. 用户体验与 API 一致性

外部调用方体验 MUST 稳定、可预测。路由、状态码、响应信封、错误消息和文档
默认保持兼容，破坏性变化必须提前说明。

- API 响应 MUST 保持统一信封：`success`、`code`、`message`、`data`、`errors`。
- `404`、`405`、`422`、`429`、`500`、`503` 的语义 MUST 明确且一致。
- 生产错误响应 MUST 不暴露内部堆栈、文件路径、连接串、密钥或原始技术错误。
- OpenAPI 文档、README、AGENTS.md 和示例配置 MUST 与实际行为同步。
- 中文项目文档和注释 SHOULD 保持简洁直接，避免教学式冗长说明。

理由：这是供前端和其他服务消费的 API 模板。一致的外部行为能降低接入成本，
也能让优化和重构不破坏现有调用方。

### IV. 性能与高可用要求

服务 MUST 保持启动快、常用路径快、依赖故障可降级。只有核心业务必须依赖
可以阻塞启动，其它外部服务默认可降级。

- 服务 SHOULD 在 30 秒内进入可接收请求状态。
- 常用请求在参考负载下 SHOULD 达到 95% 请求 1 秒内完成。
- `/health` MUST 只表示进程存活；`/ready` MUST 表达正常、降级和不可用状态。
- 非关键依赖失败 MUST 不导致整个 HTTP 服务退出。
- 中间件、日志和健康检查 MUST 避免每次请求执行不必要的重复或阻塞工作。
- 性能优化 MUST 有 smoke、load、基准或集成验证，不能只依赖主观判断。

理由：项目定位为生产后端模板。可用性和性能问题通常来自启动链路、依赖链路和
请求中间件链路，必须在治理层直接约束。

## 安全与运维约束

- Secret、令牌、密码和连接串 MUST 只来自受控配置或环境变量，不得硬编码进代码或文档示例。
- 日志、错误响应、健康检查和调试输出 MUST 对敏感信息脱敏。
- 认证、鉴权、限流、请求体限制、超时、CORS、代理信任和安全响应头 MUST 有安全默认值。
- 数据库 schema 变更 MUST 使用 Atlas 工作流，迁移文件和校验文件必须保持一致。
- CI SHOULD 覆盖测试、lint、govulncheck/CodeQL、Docker 构建和依赖更新检查。

## 开发工作流与质量门禁

- 规划阶段 MUST 在 `plan.md` 的 Constitution Check 中声明本章程如何被满足。
- 任务阶段 MUST 把测试、文档、性能和安全验证拆成可执行任务。
- 实现阶段 MUST 优先保持现有外部行为兼容；必要破坏性变化必须写入任务和交付说明。
- 合并前 SHOULD 运行 `rtk make check-all`、`rtk make coverage`、`rtk make docs`。
- 影响性能或可用性的变更 SHOULD 运行 `rtk make k6-root` 和 `rtk make k6-root-load`。
- 如果某项门禁无法运行，交付说明 MUST 写明原因、风险和替代验证方式。

## Governance

本章程优先于普通偏好和临时实现习惯。AGENTS.md、spec-kit 模板、README、CI 和任务计划
必须与本章程保持一致。

修订规则：

- MAJOR：删除原则、放宽不可协商要求或改变治理含义。
- MINOR：新增原则、章节或显著扩展约束。
- PATCH：措辞澄清、示例更新、错别字或不改变含义的修正。

任何修订 MUST：

- 更新本文件顶部 Sync Impact Report。
- 更新受影响的 `.specify/templates/` 文件和运行时指导文档。
- 使用 ISO 日期更新 `Last Amended`。
- 在后续 plan/tasks 中通过 Constitution Check 体现新约束。

**Version**: 1.0.0 | **Ratified**: 2026-05-19 | **Last Amended**: 2026-05-19
