# 项目规范指南

## 项目结构与模块组织

本项目是一个基于 Go 1.26.3 + Fiber v3 的后端。单一入口点为 `cmd/app/main.go`；HTTP 服务通过 `serve` 子命令启动，命令行（CLI）命令通过 Cobra 实现。代码采用受 Laravel 启发的设计架构：

- **`app/`**: 核心应用逻辑。
  - `Http/`: 控制器 (Controllers)、中间件 (Middleware)、请求 (Requests) 和 API 服务。
  - `Console/`: CLI 命令和计划任务。
  - `Models/`: 数据库领域模型和实体。
  - `Providers/`: 基础设施初始化（数据库、Redis、邮件、搜索等）。
  - `Services/`: 业务逻辑层。
  - `Support/`: 共享助手函数和工具。
- **`bootstrap/`**: 应用启动引导和依赖注入。
- **`configs/`**: 配置管理。
  - `yml/`: 包含实际配置数据的模块化 YAML 文件。
  - `internal/`: 用于加载、默认值和环境映射的 Go 实现。
  - `config.go`: 提供类型和加载方法的公共门面。
- **`database/`**: 迁移、种子和工厂数据。
- **`routes/`**: API 路由定义 (v1, v2 等)。
- **`docs/`**: 生成的 OpenAPI 3.1 文档（用于 Scalar UI）。
- **`tests/`**: 集中测试目录。
  - `unit/`: 隔离单元测试。
  - `integration/`: 端到端和集成测试。

## 构建、测试与开发命令

- `make init`: 安装工具、同步依赖并设置 `.env`。
- `make dev`: 启动热重载开发服务器 (Air)。
- `make run`: 直接运行服务器。
- `make test`: 运行 `tests/` 目录下的所有测试。
- `make check`: 运行格式化、Lint 检查和测试。
- `make docs`: 为 Scalar 重新生成 OpenAPI 3.1 输出。
- `make atlas-diff NAME=<name>` & `make atlas-apply`: 管理数据库迁移。

## 代码风格与命名规范

- 使用 `gofmt` 进行格式化，缩进使用 Tab。
- 使用 `%w` 进行显式的错误处理与包裹。
- 保持文件专注于单一职责。
- 通过 `.golangci.yml` 强制执行 Lint 检查 (govet, staticcheck, errcheck, gosec 等)。

## 测试指南

- **零共存原则 (Zero-Coexistence Rule)**: 所有测试文件必须放置在 `tests/` 目录下。`app/` 或 `configs/` 目录中不得存在 `*_test.go` 文件。
- **断言标准**: 所有测试必须使用 `github.com/stretchr/testify` (`assert` and `require` 包)。禁止使用原生的 `if err != nil` 或 `t.Errorf` 进行标准检查。
- **命名规范**: 使用 `TestFeature_Behavior` 或 `TestPackage_Function` 模式。
- 覆盖范围应包括中间件、控制器、队列行为、存储和响应格式化。

## 提交与合并请求指南

- 使用简洁、描述性的全小写提交信息 (例如: `fix: storage driver mapping`)。
- 合并请求 (Pull Request) 应包含更改摘要和验证结果。

## 安全与配置建议

- 严禁硬编码秘密。始终使用 `.env` 或环境变量。
- 配置加载支持 `${VAR:default}` 扩展。
- 验证所有 HTTP 和 CLI 输入；确保在 `configs/yml/security.yaml` 中正确配置了 CSRF 和 CORS。

## 应用容器与 Provider

本项目使用集中的应用容器来管理基础设施依赖。你应该优先使用 `app/Providers/providers.go` 中提供的全局访问方法。

### 全局访问

使用 `providers.App()` 访问全局 `Runtime` 实例，从而获取所有已注册的服务：

```go
rt := providers.App()
db := rt.Connection // 获取数据库连接
cfg := rt.Config    // 获取配置
```

### 可用服务 (Providers)

`Runtime` 实例提供以下服务：

| 服务 | 类型 | 说明 |
| :--- | :--- | :--- |
| `Config` | `*configs.Config` | 应用配置数据 |
| `ConfigRepo` | `configContracts.Repository` | 用于动态查找的配置仓库 |
| `Database` | `*database.Manager` | 数据库管理器 |
| `Connection` | `databaseContracts.Connection` | 主要数据库连接 (Bun DB) |
| `CacheManager` | `*cache.Manager` | 缓存管理器 |
| `Cache` | `cacheContracts.Store` | 默认缓存存储 (Redis/Memory) |
| `Auth` | `*auth.Manager` | 身份验证与会话管理器 |
| `MailManager` | `*mail.Manager` | 邮件服务管理器 |
| `EmailService` | `mailContracts.Mailer` | 默认邮件服务 |
| `Realtime` | `*realtime.Manager` | 实时通信 (WebSocket) 管理器 |
| `QueueManager` | `*queue.Manager` | 任务队列管理器 |
| `QueueService` | `queueContracts.Queue` | 默认队列服务 |
| `ScheduleManager` | `*schedule.Manager` | 任务调度管理器 |
| `ScheduleService` | `scheduleContracts.Scheduler` | 任务调度服务 |
| `SearchManager` | `*search.Manager` | 搜索引擎管理器 |
| `SearchService` | `searchContracts.Engine` | 默认搜索引擎服务 |
| `Storage` | `storageContracts.StorageManager` | 文件系统与云存储管理器 |
| `Hash` | `hashContracts.Hasher` | 密码与数据哈希服务 |
| `Notification` | `notificationContracts.Dispatcher` | 多渠道通知分发器 |
| `Translator` | `i18nContracts.Translator` | 国际化与本地化服务 |
| `Validation` | `validationContracts.Factory` | 请求验证工厂 |
| `Log` | `loggingContracts.Logger` | 应用日志服务 |
| `RateLimiter` | `ratelimiterContracts.Limiter` | API 限流服务 |

## 黄金法则

**始终在命令前添加 `rtk` 前缀**。如果 RTK 有专用过滤器，它将使用该过滤器；如果没有，它将原样传递命令。这意味着使用 RTK 始终是安全的。

@RTK.md