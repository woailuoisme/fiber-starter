# 饭盒售货机后端 API

## 项目概述

这是一个高可靠、高性能、产品级的自动饭盒售货机后端 RESTful API 应用程序。该应用采用 Golang **Fiber v3** 框架开发，完全仿照 PHP **Laravel** 框架的设计理念和架构模式，为 Flutter 前端应用提供强大的后端支持。

本项目采用了**延迟加载（Lazy Loading）**架构设计，确保外部依赖（如数据库、S3、Meilisearch、Redis、消息队列等）仅在业务逻辑实际需要时才建立连接，极大地提升了应用启动速度和资源利用率。

## 项目特点

- **Laravel 风格架构**：完全仿照 Laravel 的目录结构和设计模式 (Service/Repository/Provider)。
- **高性能 Fiber v3**：基于最新的 Fiber v3 框架，提供极致的 HTTP 处理性能。
- **延迟加载架构**：S3、Meilisearch、Database、Redis、Queue 等服务按需连接，启动零延迟。
- **RESTful API**：标准的 RESTful API 设计，易于集成。
- **完整的认证系统**：基于 JWT 的认证机制，支持 Token 刷新和黑名单。
- **多数据库支持**：通过 GORM 支持 MySQL、PostgreSQL、SQLite。
- **全文搜索引擎**：集成 Meilisearch 提供高性能搜索服务。
- **异步任务队列**：基于 Asynq (Redis) 的强大后台任务处理系统。
- **对象存储**：支持 AWS S3、MinIO、R2 等多种对象存储后端。
- **命令行工具**：内置强大的 CLI 工具（基于 Cobra），支持数据库迁移、数据填充、任务调度等。
- **Scalar 文档**：自动生成 OpenAPI/Swagger 规范，并通过 Scalar 展示 API 文档。
- **优雅的错误处理**：统一的 `apierrors` 包和错误响应格式。

## 技术栈

### 核心框架
- **Go**: 1.26+
- **Fiber**: v3 (Web Framework)
- **GORM**: ORM Framework
- **Viper**: 配置管理
- **Cobra**: 命令行工具
- **Dig**: 依赖注入容器

### 数据库与缓存
- **MySQL / PostgreSQL**: 主数据库
- **Redis**: 缓存、会话、队列后端
- **SQLite**: 本地开发/测试数据库

### 搜索与存储
- **Meilisearch**: 全文搜索引擎
- **AWS S3 / MinIO**: 对象存储

### 队列与任务
- **Asynq**: 分布式任务队列

### 工具库
- **Zap**: 高性能日志库
- **Validator**: 数据验证
- **Carbon**: 时间处理
- **Swag + Scalar**: API 规范生成与文档展示

## 架构设计

### 目录结构

```
.
├── cmd/                      # Main applications
│   ├── server/               # HTTP 服务入口
│   │   └── main.go
│   └── cli/                  # CLI 工具入口
│       └── main.go
├── app/                      # 应用核心代码
│   ├── apierrors/            # 统一错误定义 (避免与标准库 errors 冲突)
│   ├── helpers/              # 全局辅助函数
│   ├── http/                 # HTTP 层
│   │   ├── controllers/      # 控制器
│   │   ├── middleware/       # 中间件
│   │   ├── requests/         # 请求验证对象
│   │   └── resources/        # API 资源转换
│   ├── models/               # GORM 数据模型
│   ├── providers/            # 依赖注入服务提供者
│   ├── services/             # 业务逻辑服务 (S3, Queue, Search 等)
│   └── logger/               # 日志配置
├── bootstrap/                # 启动引导逻辑
├── command/                  # CLI 命令实现 (Migrate, Seed, Route 等)
├── config/                   # 配置加载与结构体（Go）
├── configs/                  # 配置文件（YAML）
│   ├── app.yaml
│   └── database.yaml
├── database/                 # 数据库相关
│   ├── migrations/           # 数据库迁移文件
│   └── seeders/              # 数据填充器
├── docs/                     # OpenAPI/Swagger 生成文件
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
- Docker & Docker Compose (推荐)
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
   # 1. 安装开发工具 (Air, Lint, Swag)
   # 2. 下载依赖 (go mod tidy)
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

本项目使用 `Makefile` 封装了常用操作：

### 开发与运行
- `make dev`: 启动热重载开发服务器
- `make run`: 直接运行应用
- `make build`: 构建二进制文件
- `make build-prod`: 构建生产环境二进制文件 (压缩体积)

### 数据库管理
- `make migrate`: 执行数据库迁移
- `make migrate-rollback`: 回滚上一次迁移
- `make seed`: 填充数据库种子数据
- `make seed-random`: 生成随机测试数据

### 队列与任务
- `go run ./cmd/cli queue:work`: 以独立进程运行 Asynq worker（用于生产/容器化部署）

### 代码质量
- `make lint`: 运行代码检查 (golangci-lint)
- `make fmt`: 格式化代码
- `make test`: 运行单元测试
- `make coverage`: 生成测试覆盖率报告

### 文档
- `make docs`: 生成 OpenAPI/Swagger 规范文件（运行后通过 `/docs` 查看 Scalar 文档）

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t fiber-starter .

# 运行容器
docker run -d -p 3000:3000 --env-file .env fiber-starter
```

## 许可证

MIT License
