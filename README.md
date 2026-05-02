# 饭盒售货机后端 API

## 项目概述

这是一个高可靠、高性能、面向生产的自动饭盒售货机后端 RESTful API。项目使用 Go 语言和 **Fiber v3** 开发，整体目录与职责组织方式尽量贴近 **Laravel 13**，为 Flutter 前端提供稳定的后端支持。

本项目采用**延迟加载（Lazy Loading）**设计，数据库、S3、Meilisearch、Redis、消息队列等外部依赖只会在实际需要时初始化，尽量减少启动耗时和资源占用。

## 项目特点

- **Laravel 风格目录**：按 Laravel 13 的命名和职责组织代码。
- **Fiber v3**：高性能 HTTP 框架。
- **延迟加载**：S3、Meilisearch、数据库、Redis、队列等按需连接。
- **RESTful API**：标准 REST 接口设计，便于前端集成。
- **JWT 认证**：支持 token 刷新、注销和黑名单。
- **数据库访问**：通过 pgx + sqlc 直连 PostgreSQL，保留 SQLite 作为本地测试数据库。
- **全文搜索**：集成 Meilisearch。
- **异步任务队列**：基于 Asynq + Redis。
- **对象存储**：基于 Fiber S3 驱动，支持 AWS S3、Garage、Cloudflare R2、阿里云 OSS 等 S3 兼容后端，bucket 需预先创建。
- **命令行工具**：基于 Cobra，支持迁移、种子、调度等。
- **Scalar 文档**：自动生成 OpenAPI 3.1，并通过 Scalar 展示。
- **统一错误处理**：统一异常类型和错误响应格式。

## 技术栈

### 核心框架
- **Go**: 1.26+
- **Fiber**: v3（Web 框架）
- **pgx + sqlc**: PostgreSQL 访问层
- **Koanf + godotenv**: 配置管理
- **Cobra**: 命令行工具
- **Dig**: 依赖注入容器

### 数据库与缓存
- **MySQL / PostgreSQL**：主数据库
- **Redis**：缓存、会话、队列后端
- **SQLite**：本地开发/测试数据库

### 搜索与存储
- **Meilisearch**：全文搜索引擎
- **AWS S3 / Garage / R2 / OSS**：对象存储

### 队列与任务
- **Asynq**：分布式任务队列

### 工具库
- **Zap**：高性能日志库
- **Validator**：数据验证
- **Carbon**：时间处理
- **OpenAPI 3.1 + Scalar**：API 规范展示

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

```
.
├── app/                      # 应用核心代码
│   ├── Console/              # CLI 命令与调度
│   │   ├── Commands/         # Cobra 命令实现
│   │   └── Kernel/           # 调度器内核
│   ├── Enums/                # 枚举
│   ├── Exceptions/           # 异常与错误码
│   ├── Http/                 # HTTP 层
│   │   ├── Controllers/      # 控制器
│   │   ├── Middleware/       # 中间件
│   │   ├── Requests/         # 请求验证对象
│   │   ├── Resources/        # API 资源转换
│   │   └── Services/         # HTTP 用例/业务服务
│   ├── I18n/                 # 国际化
│   ├── Models/               # 数据模型
│   ├── Providers/            # 依赖注入服务提供者
│   ├── Services/             # 基础设施服务（Search、Queue、Storage、Email）
│   └── Support/              # 日志、缓存、错误响应等支持层
├── bootstrap/                # 启动引导逻辑
├── config/                   # 配置加载与结构体
├── database/                 # 数据库迁移、工厂与种子
│   ├── factories/            # 模型工厂
│   ├── migrations/           # 数据库迁移文件
│   │   ├── postgres/         # PostgreSQL 迁移
│   │   └── sqlite/           # SQLite 迁移
│   └── seeders/              # 数据填充器
├── routes/                   # 路由定义
│   ├── api.go
│   ├── console.go
│   └── web.go
├── cmd/                      # 可执行入口
│   ├── server/               # HTTP 服务入口
│   │   └── main.go
│   └── cli/                  # CLI 入口
│       └── main.go
├── docs/                     # OpenAPI 3.1 生成文件
├── lang/                     # 语言文件
├── public/                   # 静态资源
├── storage/                  # 运行时存储
│   ├── logs/                 # 日志文件
│   └── private/              # 私有文件
├── tests/                    # 测试文件
├── .env.example              # 环境变量示例
├── .gitignore                # Git 忽略规则
├── Dockerfile                # Docker 构建文件
├── Makefile                  # 开发与构建命令
├── go.mod                    # Go 依赖定义
└── go.sum
```

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
   # 1. 安装开发工具（Air、Lint、sqlc、Atlas）
   # 2. 下载依赖（go mod tidy）
   # 3. 复制 .env 文件
   ```

3. **配置环境变量**
   编辑 `.env` 文件，配置数据库、Redis 等连接信息。

4. **查看或调整构建配置**
   ```bash
   make config
   # 或者直接查看 .buildconfig
   ```

5. **启动开发环境**
   ```bash
   make dev
   # 使用 Air 进行热重载开发
   ```

## 常用命令

本项目使用 `Makefile` 封装了常用操作：

### 开发与运行
- `make dev`：启动热重载开发服务器
- `make run`：直接运行应用
- `make build`：构建二进制文件
- `make build-prod`：构建生产环境二进制文件（压缩体积）
- `make config`：显示当前构建配置
- `./scripts/config.sh`：输出同样的构建配置

### 数据库管理
- `make migrate`：执行数据库迁移
- `make migrate-rollback`：回滚上一次迁移
- `make seed`：填充数据库种子数据
- `make seed-random`：生成随机测试数据

### 队列与任务
- `go run ./cmd/cli queue:work`：以独立进程运行 Asynq worker（用于生产/容器化部署）

### 代码质量
- `make lint`：运行代码检查（golangci-lint）
- `make fmt`：格式化代码
- `make test`：运行单元测试
- `make coverage`：生成测试覆盖率报告

### 文档
- `make docs`：生成 OpenAPI 3.1 规范文件（运行后通过 `/docs` 查看 Scalar 文档）

### 构建配置
- 配置文件：`.buildconfig`
- 默认项：
  - `BUILD_DIR=build`
  - `COVERAGE_DIR=coverage`
  - `SERVER_BINARY_NAME=fiber-starter`
  - `CLI_BINARY_NAME=fiber-starter-cli`
  - `APP_LOG_DIR=storage/logs`
  - `DEPLOY_DIR=deploy`
- 可用环境变量临时覆盖：
  - `BUILD_DIR=dist make build`
  - `COVERAGE_DIR=/tmp/coverage ./scripts/dev.sh test`

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t fiber-starter .

# 运行容器
docker run -d -p 8080:8080 \
  -e APP_ENV=production \
  -e APP_PORT=8080 \
  -e APP_HOST=0.0.0.0 \
  fiber-starter

# 使用 Docker Compose
docker compose up -d --build
```

GitHub Actions 会把镜像推送到 GitHub Container Registry，镜像名默认是 `ghcr.io/<owner>/<repo>`。

## 许可证

MIT License
