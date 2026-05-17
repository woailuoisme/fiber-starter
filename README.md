# 饭盒售货机后端 API

## 项目概述

这是一个高可靠、高性能、面向生产的自动饭盒售货机后端 RESTful API。项目使用 Go 语言和 **Fiber v3** 开发，整体目录与职责组织方式尽量贴近 **Laravel**，为 Flutter 前端提供稳定的后端支持。

本项目采用**延迟加载（Lazy Loading）**设计，数据库、S3、Meilisearch、Redis、消息队列等外部依赖只会在实际需要时初始化，尽量减少启动耗时和资源占用。

## 项目特点

- **Laravel 风格目录**：按 Laravel 的命名和职责组织代码。
- **Fiber v3**：高性能 HTTP 框架。
- **延迟加载**：S3、Meilisearch、数据库、Redis、队列等按需连接。
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
- **Scalar 文档**：自动生成 OpenAPI 3.1，并通过 Scalar 展示。
- **统一错误处理**：统一异常类型和错误响应格式。
- **国际化**：基于 Fiber 官方 `contrib/v3/i18n`，支持 `query lang`、Cookie 和 `Accept-Language`。

## 技术栈

### 核心框架

- **Go**: 1.26+
- **Fiber**: v3（Web 框架）
- **Bun + Atlas**: PostgreSQL / SQLite 访问层与迁移工作流
- **Koanf**: 配置管理核心，负责配置聚合与自动类型转换。
- **godotenv**: 环境加载，负责将 `.env` 文件载入进程空间。

### 数据库与缓存

- **MySQL / PostgreSQL**：主数据库
- **Rueidis + Redis**：高性能缓存、会话、队列后端
- **Ristretto**：高性能本地 L1 缓存
- **SQLite**：本地开发/测试数据库

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
- **OpenAPI 3.1 + Scalar**: API 规范展示

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
APP_FIBER_IDLE_TIMEOUT=30
APP_FIBER_READ_BUFFER_SIZE=16384
```

## 架构设计

### 目录结构

本项目代码采用受 Laravel 启发的设计架构：

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
- **`lang/`**: 语言文件（i18n）。
- **`storage/`**: 运行时存储。

### 应用容器与 Provider

本项目使用集中的应用容器来管理基础设施依赖。你应该优先使用 `app/Providers/providers.go` 中提供的全局访问方法。

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
- Make 工具

### 安装步骤

1. **克隆项目**

   ```bash
   git clone <repository-url>
   cd fiber-starter
   ```

2. **初始化项目**

   ```bash
   make init
   # 该命令会自动：
   # 1. 安装开发工具（Air、Lint、Atlas）
   # 2. 下载依赖（go mod tidy）
   # 3. 复制 .env 文件
   ```

3. **配置环境变量**
   编辑 `.env` 文件，配置数据库、Redis 等连接信息。

4. **启动开发环境**

   ```bash
   make dev
   # 使用 Air 进行热重载开发
   ```

## 常用命令

本项目使用 `Makefile` 封装了常用操作，**建议始终在命令前添加 `rtk` 前缀**：

### 开发与运行

- `rtk make dev`：启动热重载开发服务器
- `rtk make run`：直接运行应用
- `rtk make build`：构建二进制文件
- `rtk make build-prod`：构建生产环境二进制文件（压缩体积）

### 数据库管理

- `rtk make migrate`：执行数据库迁移（使用 Atlas）
- `rtk make seed`：填充数据库种子数据

### 异步队列

- `rtk go run ./cmd/app queue:work`：运行任务队列工作进程

### 代码质量

- `rtk make lint`：运行代码检查
- `rtk make fmt`：格式化代码
- `rtk make test`：运行单元测试

### 文档

- `rtk make docs`：重新生成 OpenAPI 3.1 规范（通过 `/docs` 访问 Scalar UI）

## 许可证

MIT License
