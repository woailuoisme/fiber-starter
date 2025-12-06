# 饭盒售货机后端 API

## 项目概述

这是一个高可靠、高性能、产品级的自动饭盒售货机后端 RESTful API 应用程序。该应用采用 Golang Fiber 框架开发，完全仿照 PHP Laravel 框架的设计理念和架构模式，为 Flutter 前端应用提供强大的后端支持。

## 项目特点

- **Laravel 风格架构**：完全仿照 Laravel 的目录结构和设计模式
- **高性能**：基于 Fiber 框架，性能优于传统 Go Web 框架
- **RESTful API**：标准的 RESTful API 设计，易于集成
- **完整的认证系统**：JWT 认证，支持 Token 刷新
- **多数据库支持**：支持 MySQL、PostgreSQL、SQLite
- **缓存系统**：Redis 缓存支持，提升性能
- **队列系统**：异步任务处理，支持邮件发送等
- **文件存储**：支持本地存储、MinIO、AWS S3
- **支付集成**：支持微信支付、支付宝支付
- **地理位置服务**：设备位置管理和上报
- **Swagger 文档**：自动生成 API 文档
- **优雅的错误处理**：统一的错误响应格式
- **数据验证**：请求数据自动验证
- **数据库迁移**：类似 Laravel 的迁移系统
- **数据填充**：Seeder 支持，方便开发测试

## 技术栈

### 核心框架
- **Fiber v2**：高性能 Web 框架
- **GORM**：ORM 框架，支持多种数据库
- **Viper**：配置管理
- **Cobra**：命令行工具

### 数据库与缓存
- **MySQL / PostgreSQL**：主数据库
- **Redis**：缓存和会话存储
- **BBolt**：嵌入式键值存储

### 认证与安全
- **JWT (golang-jwt/jwt)**：Token 认证
- **bcrypt**：密码加密

### 存储服务
- **MinIO**：对象存储
- **AWS S3**：云存储
- **本地文件系统**：本地存储

### 队列与任务
- **Asynq**：异步任务队列
- **Redis**：队列后端

### 工具库
- **Validator**：数据验证
- **Faker**：测试数据生成
- **Carbon**：时间处理
- **Swagger**：API 文档生成
- **Mail**：邮件发送

## 架构设计

### Laravel 风格的目录结构

```
.
├── app/                      # 应用核心代码（类似 Laravel app/）
│   ├── controllers/          # 控制器层
│   ├── models/              # 数据模型
│   ├── services/            # 业务逻辑层
│   ├── middleware/          # 中间件
│   ├── providers/           # 服务提供者（依赖注入容器）
│   ├── helpers/             # 辅助函数
│   ├── errors/              # 错误定义
│   ├── enums/               # 枚举类型
│   └── logger/              # 日志服务
├── config/                   # 配置文件（类似 Laravel config/）
│   ├── app.yaml             # 应用配置
│   ├── database.yaml        # 数据库配置
│   └── config.go            # 配置加载器
├── database/                 # 数据库相关（类似 Laravel database/）
│   ├── migrations/          # 数据库迁移
│   ├── seeders/             # 数据填充
│   └── connection.go        # 数据库连接
├── routes/                   # 路由定义（类似 Laravel routes/）
│   ├── api.go               # API 路由
│   └── web.go               # Web 路由
├── storage/                  # 存储目录（类似 Laravel storage/）
│   ├── logs/                # 日志文件
│   └── cache/               # 缓存文件
├── public/                   # 公共资源（类似 Laravel public/）
├── resources/                # 资源文件（类似 Laravel resources/）
├── tests/                    # 测试文件
├── docs/                     # API 文档
├── command/                  # 命令行工具（类似 Laravel artisan）
├── cmd/                      # 应用入口
├── .env                      # 环境变量
├── .env.example             # 环境变量示例
├── Makefile                 # Make 命令
├── Taskfile.yml             # Task 命令
└── main.go                  # 主入口文件
```

### 架构模式

采用 **MVC + Service Layer** 架构：

- **Model**：数据模型层，使用 GORM ORM
- **Controller**：控制器层，处理 HTTP 请求
- **Service**：业务逻辑层，封装复杂业务逻辑
- **Middleware**：中间件层，处理认证、验证、日志等
- **Provider**：服务提供者，依赖注入容器（使用 dig）

## 核心功能模块

### 1. 用户认证系统
- 用户注册
- 用户登录
- JWT Token 认证
- Token 刷新
- 密码重置
- 邮箱验证

### 2. 城市管理
- 城市列表
- 城市详情
- 城市创建/更新/删除
- 城市下的设备管理

### 3. 设备管理
- 设备列表（支持按城市筛选）
- 设备详情
- 设备状态监控
- 设备位置上报
- 设备故障管理
- 设备库存管理

### 4. 产品管理
- 产品分类
- 产品列表（支持按设备筛选）
- 产品详情
- 产品库存管理
- 产品价格管理
- 产品图片管理

### 5. 订单系统
- 订单创建
- 订单支付
- 订单状态管理
- 订单历史查询
- 订单退款

### 6. 支付系统
- 微信支付集成
- 支付宝支付集成
- 支付回调处理
- 支付状态查询
- 交易记录

### 7. 文件存储
- 图片上传
- 文件管理
- 多存储驱动支持（本地/MinIO/S3）

### 8. 通知系统
- 邮件通知
- 队列异步发送
- 通知模板管理

## 数据模型设计

### 核心实体关系

```
City (城市)
  ├── has many Devices (设备)
  
Device (设备)
  ├── belongs to City (城市)
  ├── has many Products (产品)
  └── has many Orders (订单)
  
Product (产品)
  ├── belongs to Device (设备)
  ├── belongs to Category (分类)
  └── has many OrderItems (订单项)
  
User (用户)
  └── has many Orders (订单)
  
Order (订单)
  ├── belongs to User (用户)
  ├── belongs to Device (设备)
  ├── has many OrderItems (订单项)
  └── has one Payment (支付)
  
Payment (支付)
  └── belongs to Order (订单)
```

## 安装指南

### 环境要求

- Go 1.24 或更高版本
- MySQL 8.0+ 或 PostgreSQL 13+
- Redis 6.0+
- Make 或 Task（可选）

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd fiber-starter
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，配置数据库、Redis 等
   ```

4. **配置数据库**
   编辑 `config/database.yaml` 文件，配置数据库连接信息

5. **运行数据库迁移**
   ```bash
   go run main.go migrate:run
   ```

6. **运行数据填充（可选）**
   ```bash
   go run main.go db:seed
   ```

7. **生成 JWT 密钥**
   ```bash
   go run main.go jwt:generate
   ```

## 运行指南

### 开发模式

使用 Air 热重载：
```bash
air
```

或使用 Make：
```bash
make dev
```

或使用 Task：
```bash
task dev
```

### 生产模式

```bash
go build -o app main.go
./app
```

### 使用 Docker

```bash
docker-compose up -d
```

## 命令行工具

类似 Laravel Artisan 的命令行工具：

```bash
# 运行迁移
go run main.go migrate:run

# 回滚迁移
go run main.go migrate:rollback

# 运行填充
go run main.go db:seed

# 生成 JWT 密钥
go run main.go jwt:generate

# 查看所有命令
go run main.go --help
```

## API 文档

启动应用后，访问 Swagger 文档：

```
http://localhost:3000/swagger/index.html
```

### API 端点示例

#### 认证相关
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/refresh` - 刷新 Token
- `POST /api/auth/logout` - 用户登出

#### 城市管理
- `GET /api/cities` - 获取城市列表
- `GET /api/cities/:id` - 获取城市详情
- `POST /api/cities` - 创建城市（需要管理员权限）
- `PUT /api/cities/:id` - 更新城市（需要管理员权限）
- `DELETE /api/cities/:id` - 删除城市（需要管理员权限）

#### 设备管理
- `GET /api/devices` - 获取设备列表
- `GET /api/devices/:id` - 获取设备详情
- `GET /api/cities/:cityId/devices` - 获取指定城市的设备
- `POST /api/devices/:id/location` - 上报设备位置

#### 产品管理
- `GET /api/products` - 获取产品列表
- `GET /api/products/:id` - 获取产品详情
- `GET /api/devices/:deviceId/products` - 获取指定设备的产品

#### 订单管理
- `POST /api/orders` - 创建订单
- `GET /api/orders` - 获取订单列表
- `GET /api/orders/:id` - 获取订单详情
- `POST /api/orders/:id/pay` - 支付订单

#### 文件上传
- `POST /api/storage/upload` - 上传文件

## 配置说明

### 应用配置 (config/app.yaml)

```yaml
app:
  name: "饭盒售货机"
  env: "development"
  debug: true
  port: "3000"
  host: "0.0.0.0"
  timezone: "Asia/Shanghai"
  url: "http://localhost:3000"

jwt:
  secret: "${JWT_SECRET}"
  expiration_time: 3600
  refresh_time: 604800
  expire_hours: 24
  issuer: "lunchbox-vending"
```

### 数据库配置 (config/database.yaml)

```yaml
default: mysql

connections:
  mysql:
    driver: mysql
    host: "${DB_HOST:localhost}"
    port: "${DB_PORT:3306}"
    database: "${DB_DATABASE:lunchbox_vending}"
    username: "${DB_USERNAME:root}"
    password: "${DB_PASSWORD:}"
    charset: utf8mb4
    collation: utf8mb4_unicode_ci
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./app/services

# 运行测试并显示覆盖率
go test -cover ./...
```

## 部署

### 使用脚本部署

```bash
./scripts/deploy.sh
```

### 手动部署

1. 构建应用
   ```bash
   go build -o app main.go
   ```

2. 上传到服务器

3. 配置环境变量

4. 运行迁移
   ```bash
   ./app migrate:run
   ```

5. 启动应用
   ```bash
   ./app
   ```

## 性能优化

- 使用 Redis 缓存热点数据
- 数据库查询优化和索引
- 使用连接池管理数据库连接
- 静态资源 CDN 加速
- API 响应压缩
- 合理使用队列处理耗时任务

## 安全性

- JWT Token 认证
- 密码 bcrypt 加密
- SQL 注入防护（GORM ORM）
- XSS 防护
- CSRF 防护
- 请求频率限制
- 敏感数据加密存储

## 监控与日志

- 结构化日志记录
- 错误追踪
- 性能监控
- API 访问日志

## 开发规范

### 代码风格

遵循 Go 官方代码规范和 Laravel 设计理念：

- 使用 `gofmt` 格式化代码
- 遵循 Go 命名规范
- 控制器方法命名：`Index`, `Show`, `Store`, `Update`, `Destroy`
- 服务层封装复杂业务逻辑
- 使用依赖注入

### Git 提交规范

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
test: 测试相关
chore: 构建/工具链相关
```

## 常见问题

### 1. 数据库连接失败
检查 `.env` 和 `config/database.yaml` 中的数据库配置

### 2. Redis 连接失败
确保 Redis 服务已启动，检查连接配置

### 3. JWT Token 无效
重新生成 JWT 密钥：`go run main.go jwt:generate`

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

- 项目地址：[GitHub Repository]
- 问题反馈：[GitHub Issues]

---

**注意**：这是一个产品级项目，请在生产环境中务必：
- 修改默认的 JWT 密钥
- 使用强密码
- 启用 HTTPS
- 配置防火墙
- 定期备份数据库
- 监控系统运行状态
