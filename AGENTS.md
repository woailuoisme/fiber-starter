# 项目规范指南

## 项目结构与模块组织

本项目是一个基于 Go 1.26.3 + Fiber v3 的后端。单一入口点为 `cmd/app/main.go`；HTTP 服务通过 `serve` 子命令启动，命令行（CLI）命令通过 Cobra 实现。代码采用受 Laravel 启发的高度自治、**Feature-First (特性优先)** 的现代 Go 架构模式：

- **`internal/`**: 核心应用逻辑目录。
  - **`features/`**: 业务特性切片目录。每个业务模块自成一体，独立自治。每个特性通常包含：
    - `routes.go`: 该特性的 HTTP 路由注册定义。
    - `controllers/` 或业务文件: 接收请求、处理响应及 HTTP 控制器逻辑。
    - `services/` 或领域文件: 承载该特性的核心领域与业务逻辑。
    - `repository.go`: 封装针对数据库/数据持久层的 CRUD 交互。
    - `requests.go`: 承载 HTTP 请求体与 DTO 的定义及参数校验规则。
    - `models/`: 属于该特性的数据库领域模型与实体定义。
  - **`providers/`**: 统一的基础设施服务适配器层。集中提供对缓存、数据库、搜索、存储、邮件、队列等的生命周期和初始化管理。
  - **`bootstrap/`**: 应用的启动引导。包含全局容器拼装、依赖注入上下文搭建与 Fiber 引擎的底层装载。
  - **`common/`**: 共享核心类库。如全局异常拦截器、API 统一响应格式包装、公用路由中间件等。
  - **`console/`**: Cobra CLI 命令行控制台指令集以及 Kernel 级别的定时计划任务管理器。
  - **`support/`**: 纯辅助性无状态助手类库与高层全局 Facade（门面）。如 `support.Logger`、`utils.go`。
- **`configs/`**: 模块化配置项管理。
  - `yml/`: 存放按环境隔离、模块解耦的纯 YAML 配置数据文件。
  - `internal/`: 封装加载逻辑，用于处理环境变量注入 `${VAR:default}` 与默认值映射。
  - `config.go`: 提供全局配置实体类型导出的公共门面。
- **`database/`**: 数据库迁移文件 (Atlas CLI 管理)、数据种子 (Seeders) 和测试数据工厂。
- **`routes/`**: 顶层全局核心路由配置入口（绑定 V1, V2 等 API 路由）。
- **`docs/`**: 由 API 自动生成的 OpenAPI 3.1 协议规范文件（用于提供 Scalar UI 调试界面）。
- **`tests/`**: 全局集中测试中心。
  - `unit/`: 独立、零外部依赖的单元测试。
  - `integration/`: 端到端及需要数据库等外部介质的集成测试。
  - `contract/`: 严格的代码规范检测合约测试（如规范静态检查）。

## 构建、测试与开发命令

- `rtk make init`: 初始化开发环境，安装工具集，同步 Go modules 依赖并建立本地 `.env`。
- `rtk make dev`: 启动 Air 热重载开发服务器。
- `rtk make run`: 生产环境模式下直接编译并运行 HTTP 实例。
- `rtk make test`: 运行 `tests/` 目录下的所有自动化测试。
- `rtk make check`: 自动化格式化校验、严格静态代码 Lint 检查与全量测试。
- `rtk make check-all`: 完整质量门禁，用于实现完成前最终验证。
- `rtk make coverage`: 运行覆盖率门禁并生成覆盖率报告。
- `rtk make docs`: 重新生成 OpenAPI 3.1 定义并导出，刷新 Scalar 文档界面。
- `rtk make k6-root` / `rtk make k6-root-load`: 对根路径执行 smoke/load 性能验证；运行前需确保服务监听在 `BASE_URL`，默认 `http://localhost:3300`。
- `rtk make atlas-diff NAME=<name>` & `rtk make atlas-apply`: 通过 Atlas 控制和应用数据库迁移 Schema 的演进。
- `./artisan <command>`: Laravel Artisan 风格短入口，例如 `./artisan jwt:secret`、`./artisan serve`、`./artisan queue:work`。
- `rtk make artisan CMD="jwt:secret"`: 通过 Makefile 运行任意 CLI 命令。
- `rtk make jwt`: 生成 32 字节随机 JWT secret，并替换 `.env` 中的 `JWT_SECRET`。

## 代码风格与命名规范

- **严格的格式化**: 使用 `gofmt` (或 `gofumpt`) 格式化所有代码，强制缩进使用 Tab。
- **显式错误链传递**: 绝对禁止忽略任何 `error`，底层错误向上传递时必须使用 `%w` 进行显式包裹包装。
- **单一职责专注度**: 单个文件只做一件事情，路由、控制器与核心服务要做到完美解耦。
- **API 空切片 JSON 序列化规范**: 在 API 序列化或 Marshaling 流程中，返回数据集合的切片切忌返回 `nil`，必须将其初始化为 concrete 的空切片 `[]Type{}`。这样可以确保输出给前端消费者的 JSON 是符合预期的空数组 `[]`，从而彻底杜绝 JavaScript/Kotlin 客户端等消费解析时发生 runtime null pointer 崩溃。
- **静态 Lint 检测**: 团队通过 `.golangci.yml` 强制审查所有 MR (govet, staticcheck, errcheck, gosec 等)，本地运行 `make check` 无警告方能提交。

## 测试指南

- **零共存原则 (Zero-Coexistence Rule)**: 所有测试文件必须放置在 `tests/` 目录下。`internal/`、`configs/` 等主要代码目录中不得存在任何 `*_test.go` 测试文件，确保生产代码与测试隔离。
- **断言规范**: 强制全部使用 `github.com/stretchr/testify` 包下的 `assert` 和 `require` 进行状态验证，严禁在业务逻辑校验中使用手写 `if err != nil { t.Errorf(...) }` 行为。
- **命名语义约束**: 单个测试用例命名推荐使用 `TestFeature_Behavior`（集成/行为）或 `TestPackage_Function`（单元）的模式，清晰自明。
- **测试环境桩化隔离**: 单元测试如涉及到 Gotify、Telegram、Mail 等第三方有网络边界或副作用的服务时，必须设计 Mock Stub 或者是利用本地 `httptest.NewServer` 实现完美捕获，禁止在测试期与真实的外网通道建立套接字物理连接。

## 应用容器与 Provider

本项目使用集中的应用容器管理基础设施。你应该优先使用 `internal/providers/providers.go` 中提供的全局访问方法。

### 就绪、降级与安全默认值

- `/health` 只表示进程存活；`/ready` 返回 `ok`、`degraded` 或 `fail`，只有关键依赖失败时返回 `503`。
- 默认数据库为关键依赖；缓存、邮件、队列、搜索、存储和实时通信为可降级依赖，可用 `SERVICE_*_CRITICAL` 环境变量覆盖。
- `JWT_SECRET`、API key、密码、token、连接串等敏感值必须来自环境变量或密钥管理系统，源码默认值不得包含真实 secret。
- 需要生成本地 JWT secret 时使用 `./artisan jwt:secret` 或 `rtk make jwt`，不要手写弱 secret；不要新增重复的 `jwt:generate` 类命令入口。
- 日志与 readiness 错误必须使用 `internal/support/redaction.go` 的脱敏辅助函数。
- 非关键 provider 构建或健康检查失败时，服务应保持可启动并在 `/ready` 中报告 `degraded`；配置错误和关键依赖失败不能被静默吞掉。
- HTTP 成功、错误和分页响应必须保持 `success`、`code`、`message`、`data`、`errors` 的公开信封；API 切片响应不得输出 JSON `null`。
- 不要在项目脚本、Makefile、文档或测试中随意写入本机专用绝对路径或缓存路径（如 `GOCACHE`、`/private/tmp/...`、用户目录）。需要项目级缓存时优先使用仓库已有且被 `.gitignore` 忽略的 `.cache/` 约定；只有用户明确要求、项目已有配置约定或命令运行时临时传参时才允许指定路径。

### 全局访问

使用 `providers.App()` 访问全局 `Runtime` 实例，从而获取所有已注册的服务：

```go
rt := providers.App()
db := rt.Connection // 获取数据库连接
cfg := rt.Config    // 获取配置
```

### 可用服务 (Providers)

`Runtime` 实例提供以下基础设施服务：

| 服务 | 类型 | 说明 |
| :--- | :--- | :--- |
| `Config` | `*configs.Config` | 全局应用配置信息定义 |
| `ConfigRepo` | `configContracts.Repository` | 用于动态读取/重载/查询的配置仓储 |
| `Database` | `databaseContracts.Manager` | 基础物理数据库链接管理器 |
| `Connection` | `databaseContracts.Connection` | 核心主要数据库连接实体 (Bun DB) |
| `CacheManager` | `cacheContracts.Manager` | 缓存多驱动管理器 |
| `Cache` | `cacheContracts.Store` | 当前生效的默认缓存存储驱动 (Redis/Memory) |
| `Auth` | `authContracts.Manager` | 身份验证鉴权及会话状态管理器 |
| `MailManager` | `mailContracts.Manager` | 邮件物理连接管理器 |
| `EmailService` | `mailContracts.Mailer` | 默认可用邮件分发服务驱动 |
| `Realtime` | `realtimeContracts.Manager` | 基于实时通信 (WebSocket) 协议连接的管理器 |
| `QueueManager` | `queueContracts.Manager` | 高性能后台异步队列任务管理器 |
| `QueueService` | `queueContracts.Queue` | 默认队列分派投递服务驱动 |
| `ScheduleManager` | `scheduleContracts.Manager` | 系统秒级定时计划任务管理器 |
| `ScheduleService` | `scheduleContracts.Scheduler` | 定时计划注册与调度服务驱动 |
| `SearchManager` | `searchContracts.Manager` | 搜索引擎连接管理器 |
| `SearchService` | `searchContracts.Engine` | 默认可用的搜索引擎服务驱动 (Meilisearch) |
| `Storage` | `storageContracts.StorageManager` | 全局物理云存储与本地文件系统统一管理器 |
| `Hash` | `hashContracts.Hasher` | 密码哈希与加密算法服务 facade |
| `Notification` | `notificationContracts.Dispatcher` | 负责多渠道分发的系统级通知总线 |
| `Translator` | `i18nContracts.Translator` | 多语言国际化本地翻译与动态校验适配服务 |
| `Validation` | `validationContracts.Factory` | 高性能请求体参数通用校验工厂服务 |
| `Log` | `loggingContracts.Logger` | 统合日志输出管理组件服务 |
| `RateLimiter` | `ratelimiterContracts.Limiter` | 系统级 HTTP 速率限制器拦截防刷服务 |

### 日志系统高层测试桩挂载

当我们需要在类似 `request_id_logging_test.go` 等孤立集成测试（此时 `appctx.App()` 为 `nil`，容器未加载）中验证底层的日志记录行为时，可通过在测试 setup 中动态桩化挂载 `logging.DefaultLogger` 来拦截全局 Facade 事件：

```go
core, observed := observer.New(zapcore.DebugLevel)
prevLogger := logging.DefaultLogger
logging.DefaultLogger = zap.New(core)
t.Cleanup(func() {
    logging.DefaultLogger = prevLogger
})
```

## 黄金法则

**始终在命令前添加 `rtk` 前缀**。如果 RTK 有专用过滤器，它将使用该过滤器；如果没有，它将原样传递命令。这意味着使用 RTK 始终是安全的。

@RTK.md

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
file:///Users/seaside/Projects/go/fiber-template/specs/001-fiber-best-practices/plan.md
<!-- SPECKIT END -->
