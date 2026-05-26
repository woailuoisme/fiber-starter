# 饭盒售货机后端 API

## 项目概述

这是一个高可靠、高性能、面向生产的自动饭盒售货机后端 RESTful API。项目使用 Go 语言和 **Fiber v3** 开发，整体目录与职责组织方式尽量贴近 **Laravel**，为 Flutter 前端提供稳定的后端支持。

本项目采用**延迟加载（Lazy Loading）**设计，数据库、S3、Meilisearch、Redis、消息队列等外部依赖只会在实际需要时初始化，尽量减少启动耗时和资源占用。

## 项目特点

- **Laravel 风格目录**：按 Laravel 的命名和职责组织代码。
- **Fiber v3**：高性能 HTTP 框架。
- **稳定启动与降级**：非关键依赖异常时服务可继续启动，并通过 `/ready` 返回 `degraded`。
- **延迟加载**：S3、Meilisearch、数据库、Redis、队列等按需连接，避免启动期阻塞。
- **RESTful API**：标准 REST 接口设计，便于前端集成。
- **JWT 认证**：支持 token 刷新、注销和黑名单。
- **数据库访问**：通过 Bun ORM 访问 PostgreSQL 和 SQLite，Atlas 负责 schema diff 与迁移生成。
- **高性能缓存**：基于 **Rueidis** (L2) + **Ristretto** (L1) 的二级缓存架构，支持 Redis 客户端缓存。
- **全文搜索**：集成 Meilisearch。
- **异步任务队列**：基于 Asynq + Redis。
- **对象存储**：基于 Fiber S3 驱动，支持 AWS S3、Garage、Cloudflare R2、阿里云 OSS 等 S3 兼容后端。
- **Excel 导入导出**：基于 **Excelize v2** 实现高效的表格数据处理。
- **可观测性**：集成 **OpenTelemetry (OTEL)**，支持分布式追踪与指标监控。
- **命令行工具**：基于 Cobra，支持迁移、种子、调度等。
- **Swagger UI 文档**：通过 `swag init` 生成 Swagger 2.0 JSON，并由官方 Fiber contrib Swagger UI 展示。
- **统一错误处理**：统一异常类型和错误响应格式。
- **安全默认值**：默认关闭 Debug，源码不提供 JWT secret，日志和 readiness 错误会脱敏。
- **国际化**：基于 Fiber 官方 `contrib/v3/i18n`，支持 `query lang`、Cookie 和 `Accept-Language`。

## 项目治理

项目章程位于 `.specify/memory/constitution.md`，它定义了代码质量、测试标准、
API 一致性、性能与高可用、安全运维等不可协商规则。所有 spec-kit 的
`plan.md` 和 `tasks.md` 都必须显式通过这些门禁；实现完成后需说明已运行的
`rtk` 验证命令，以及任何无法运行检查的原因和风险。

## 技术栈

### 核心框架

- **Go**: 1.26+
- **Fiber**: v3（Web 框架）
- **Bun + Atlas**: PostgreSQL / SQLite 访问层与迁移工作流
- **Koanf**: 配置管理核心，负责配置聚合与自动类型转换。
- **godotenv**: 环境加载，负责将 `.env` 文件载入进程空间。

### 数据库与缓存

- **PostgreSQL / SQLite**：主数据库与本地开发/测试数据库
- **Rueidis + Redis**：高性能缓存、会话、队列后端
- **Ristretto**：高性能本地 L1 缓存

### 搜索与存储

- **Meilisearch**：全文搜索引擎
- **AWS S3 / Garage / R2 / OSS**：对象存储

### 队列与任务

- **Asynq**：分布式任务队列

### 观测与工具

- **OpenTelemetry**：分布式追踪与监控 (Otlp/gRPC)
- **Zap**：高性能日志库
- **Excelize v2**：Excel 导入导出
- **Resend**：邮件发送 SDK
- **Validator**: 数据验证
- **Carbon**: 时间处理
- **Swagger UI + Swagger 2.0 JSON**: API 规范展示

## Fiber 启动配置

Fiber 的启动参数统一通过配置文件和环境变量管理，当前可配置项主要包括：

- `APP_FIBER_PREFORK`：是否开启 Prefork 多进程模式，启动时映射到 `fiber.ListenConfig.EnablePrefork`
- `APP_FIBER_SERVER_HEADER`：响应头 `Server`
- `APP_FIBER_BODY_LIMIT`：请求体大小上限
- `APP_FIBER_CONCURRENCY`：最大并发连接数
- `APP_FIBER_READ_BUFFER_SIZE`：读取缓冲区大小
- `APP_FIBER_READ_TIMEOUT` / `APP_FIBER_WRITE_TIMEOUT` / `APP_FIBER_IDLE_TIMEOUT`：请求读写与空闲超时
- `APP_FIBER_TRUST_PROXY` / `APP_FIBER_PROXY_HEADER`：代理场景下的真实客户端 IP 识别
- `APP_FIBER_STREAM_REQUEST_BODY`：大请求体流式读取
- `APP_FIBER_IMMUTABLE`：将上下文返回值设为不可变

单机部署时，通常建议将 `APP_FIBER_PREFORK=true`，让 Fiber 通过多进程更充分利用多核 CPU；如果你已经在 Kubernetes、Docker Compose 多副本或负载均衡后面部署，也可以保持 `false`，交给外层做水平扩容。

配置示例：

```env
APP_FIBER_PREFORK=true
APP_FIBER_READ_TIMEOUT=30
APP_FIBER_WRITE_TIMEOUT=30
APP_FIBER_IDLE_TIMEOUT=120
APP_FIBER_READ_BUFFER_SIZE=16384
```

## 健康检查与降级策略

- `/health`：只表示进程存活，适合容器 liveness probe。
- `/ready`：聚合数据库、缓存、邮件、队列、搜索、存储等依赖状态，适合 readiness probe。
- 默认数据库为关键依赖；缓存、邮件、队列、搜索、存储和实时通信为可降级依赖。
- `/ready` 返回 `ok` 或 `degraded` 时 HTTP 状态码为 `200`；关键依赖失败时返回 `fail` 和 `503`。
- 可通过 `SERVICE_DATABASE_CRITICAL`、`SERVICE_CACHE_CRITICAL`、`SERVICE_MAIL_CRITICAL`、`SERVICE_QUEUE_CRITICAL`、`SERVICE_SEARCH_CRITICAL`、`SERVICE_STORAGE_CRITICAL`、`SERVICE_REALTIME_CRITICAL` 调整依赖关键性。

示例：

```env
SERVICE_DATABASE_CRITICAL=true
SERVICE_CACHE_CRITICAL=false
SERVICE_MAIL_CRITICAL=false
SERVICE_QUEUE_CRITICAL=false
SERVICE_SEARCH_CRITICAL=false
SERVICE_STORAGE_CRITICAL=false
SERVICE_REALTIME_CRITICAL=false
```

## 安全配置

- 生产环境必须保持 `APP_DEBUG=false`。
- `JWT_SECRET`、邮件 API key、对象存储密钥、支付密钥、数据库密码等必须通过环境变量或密钥管理系统注入。
- 源码配置中的 secret 默认值应为空字符串，不允许提交真实 token、password、API key 或连接串。
- HTTP 访问日志、405 诊断日志和 readiness 依赖错误会对 token、password、API key、Authorization、Cookie 和连接串做脱敏。
- 默认请求体限制为 4 MiB，读超时 30 秒，写超时 30 秒，空闲超时 120 秒，默认不信任代理。

## 架构设计

### 目录结构

本项目代码采用受 Laravel 启发的 Feature-First 架构：

- **`cmd/app/`**: 单一入口点，HTTP 服务通过 `serve` 子命令启动。
- **`internal/features/`**: 业务特性切片，每个特性自包含路由、请求 DTO、控制器/handler、服务、仓储和模型。
- **`internal/providers/`**: 基础设施 provider，管理数据库、缓存、邮件、队列、搜索、存储、日志、验证、限流等生命周期。
- **`internal/bootstrap/`**: 应用启动引导、HTTP app 创建、路由注册和运行器。
- **`internal/common/`**: 共享异常、请求绑定、全局中间件和队列基础设施。
- **`internal/support/`**: 无状态辅助函数、Facade、响应信封、错误映射、脱敏和健康聚合。
- **`internal/console/`**: Cobra CLI 命令和计划任务。
- **`configs/`**: 配置管理。
  - `yml/`: 包含实际配置数据的模块化 YAML 文件。
  - `internal/`: 用于加载、默认值和环境映射的 Go 实现。
  - `config.go`: 提供类型和加载方法的公共门面。
- **`database/`**: 迁移、种子和工厂数据。
- **`docs/`**: 生成的 Swagger 2.0 文档（用于 Swagger UI）。
- **`tests/`**: 集中测试目录。
  - `unit/`: 隔离单元测试。
  - `integration/`: 端到端和集成测试。
  - `contract/`: HTTP/readiness 等外部契约测试。
- **`lang/`**: 语言文件（i18n）。
- **`storage/`**: 运行时存储。

### 应用容器与 Provider

本项目使用集中的应用容器来管理基础设施依赖。你应该优先使用 `internal/providers/providers.go` 中提供的全局访问方法。

#### 全局访问

使用 `providers.App()` 访问全局 `Runtime` 实例，从而获取所有已注册的服务：

```go
rt := providers.App()
db := rt.Connection // 获取数据库连接
cfg := rt.Config    // 获取配置
```

#### 可用服务 (Providers)

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

## 快速开始

### 前置要求

- Go 1.26+
- Docker & Docker Compose（推荐）
- Just 命令运行器

### 安装步骤

1. **克隆项目**

   ```bash
   git clone <repository-url>
   cd lfiber
   ```

2. **初始化项目**

   ```bash
   rtk just init
   # 该命令会自动：
   # 1. 安装开发工具（Air、Lint、Atlas）
   # 2. 下载依赖（go mod tidy）
   # 3. 复制 .env 文件
   ```

3. **配置环境变量**
   编辑 `.env` 文件，配置数据库、Redis 等连接信息。`JWT_SECRET`、邮件 API key、对象存储密钥等敏感值必须通过环境变量或密钥管理系统注入，禁止写入源码；生产环境保持 `APP_DEBUG=false`。

   生成本地 JWT secret：

   ```bash
   ./artisan jwt:generate
   # 或指定环境文件
   ./artisan jwt:generate --env .env.local
   ```

4. **配置依赖关键性**
   默认数据库为关键依赖，缓存、邮件、队列、搜索、存储和实时通信为可降级依赖。可通过 `SERVICE_DATABASE_CRITICAL`、`SERVICE_CACHE_CRITICAL` 等环境变量调整 `/ready` 的 `ok`、`degraded`、`fail` 行为。

5. **启动开发环境**

   ```bash
   rtk just dev
   # 使用 Air 进行热重载开发
   ```

## 常用命令

本项目使用 `justfile` 封装了常用操作，**建议始终在命令前添加 `rtk` 前缀**：

### 开发与运行

- `rtk just dev`：启动热重载开发服务器
- `rtk just run`：直接运行应用
- `rtk just build`：构建二进制文件
- `rtk just build-prod`：构建生产环境二进制文件（压缩体积）

### 数据库管理

- `rtk just migrate`：执行数据库迁移（使用 Atlas）
- `rtk just seed`：填充数据库种子数据
- `./artisan db:reset --force`：跳过交互确认并重置数据库，适合受控脚本环境
- `./artisan db:fresh --no-interaction`：禁用交互输入；未显式 `--force` 时会安全取消破坏性操作

### 异步队列

- `./artisan queue:work`：运行任务队列工作进程

### 密钥生成

- `./artisan jwt:generate`：生成 32 字节随机 JWT secret，并替换 `.env` 中的 `JWT_SECRET`
- `./artisan jwt:generate --env .env.local`：替换指定环境文件中的 `JWT_SECRET`
- `rtk just jwt`：通过 justfile 执行同一命令
- `rtk just artisan route:list`：通过 justfile 运行任意 CLI 命令
- JWT secret 只保留 `jwt:generate` 一个命令入口，不再提供重复的 `jwt:secret` 命令。

### 代码质量

- `rtk just lint`：运行代码检查
- `rtk just fmt`：格式化代码
- `rtk just test`：运行单元测试
- `rtk just check`：运行格式、静态检查与测试门禁
- `rtk just check-all`：运行完整质量门禁
- `rtk just coverage`：运行覆盖率门禁

### 文档

- `rtk just docs`：重新生成 Swagger 2.0 规范（通过 `/docs` 访问 Swagger UI）

### 性能验证

运行 k6 前先启动服务，默认 `BASE_URL` 为 `http://localhost:3300`：

```bash
APP_PORT=3300 DB_CONNECTION=sqlite DB_SQLITE_DATABASE=/tmp/fiber-template-k6.sqlite CACHE_DRIVER=memory STORAGE_DRIVER=local STORAGE_LOCAL_ROOT=/tmp/fiber-template-storage STORAGE_LOCAL_URL=/storage I18N_LANGUAGE_DIR=$(pwd)/lang ./artisan serve
rtk just k6-root
rtk just k6-root-load
```

k6 默认门禁：错误率 `< 1%`，常用路径 p95 `< 1s`。

## 许可证

MIT License
