# 配置系统说明

## 概述

本项目采用 Laravel 风格的配置系统，使用 Viper 加载 YAML 配置文件，并支持环境变量覆盖。

## 配置文件

### app.yaml
应用程序主配置文件，包含：
- 应用基础配置（名称、环境、端口等）
- JWT 认证配置
- Redis 配置
- 日志配置
- 缓存配置
- 邮件配置
- 队列配置
- 存储配置
- WebSocket 配置
- 支付配置
- 业务配置
- 安全配置

### database.yaml
数据库配置文件，支持：
- PostgreSQL（主数据库）
- MySQL（备用）
- SQLite（开发/测试）
- 连接池配置
- 迁移配置
- 填充配置

## 环境变量

### 优先级
环境变量 > YAML 配置文件 > 默认值

### 使用方式

1. 复制 `.env.example` 到 `.env`
```bash
cp .env.example .env
```

2. 修改 `.env` 文件中的配置项

3. 在 YAML 文件中使用环境变量占位符：
```yaml
database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:5432}
```

格式：`${环境变量名:默认值}`

## 配置加载

### 初始化配置
```go
import "fiber-starter/config"

func main() {
    // 加载配置
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用配置
    fmt.Println(cfg.App.Name)
}
```

### 访问配置
```go
// 通过全局配置实例
config.GlobalConfig.App.Name

// 通过 Viper
config.GetString("app.name")
```

### 环境变量辅助函数
```go
// 获取字符串
dbHost := config.GetEnvString("DB_HOST", "localhost")

// 获取整数
dbPort := config.GetEnvInt("DB_PORT", 5432)

// 获取布尔值
debug := config.GetEnvBool("APP_DEBUG", true)

// 获取必需的环境变量（不存在会 panic）
jwtSecret := config.MustGetEnv("JWT_SECRET")
```

## 配置结构

### 应用配置
```go
type AppConfig struct {
    Name     string
    Env      string
    Debug    bool
    Port     string
    Host     string
    Timezone string
    URL      string
}
```

### 数据库配置
```go
type DatabaseConfig struct {
    Default     string
    Connections map[string]DBConnection
    Pool        DBPoolConfig
    Migrations  DBMigrationConfig
    Seeders     DBSeederConfig
}
```

### JWT 配置
```go
type JWTConfig struct {
    Secret         string
    ExpirationTime int
    RefreshTime    int
    Issuer         string
}
```

### Redis 配置
```go
type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}
```

### 存储配置
```go
type StorageConfig struct {
    Driver string
    MinIO  *MinIOStorageConfig
    S3     *S3StorageConfig
}
```

### WebSocket 配置
```go
type WebSocketConfig struct {
    Port              string
    Path              string
    HeartbeatInterval int
}
```

### 支付配置
```go
type PaymentConfig struct {
    Wechat WechatPaymentConfig
    Alipay AlipayPaymentConfig
}
```

### 业务配置
```go
type BusinessConfig struct {
    Order  OrderConfig
    Device DeviceConfig
}
```

### 安全配置
```go
type SecurityConfig struct {
    CORS      CORSConfig
    RateLimit RateLimitConfig
}
```

## 最佳实践

1. **敏感信息**：永远不要在 YAML 文件中硬编码敏感信息（密码、密钥等），使用环境变量
2. **默认值**：为所有配置项提供合理的默认值
3. **环境区分**：使用 `APP_ENV` 区分不同环境（local, development, staging, production）
4. **配置验证**：在应用启动时验证必需的配置项
5. **文档更新**：修改配置结构时同步更新 `.env.example` 和文档

## 配置示例

### 开发环境
```env
APP_ENV=local
APP_DEBUG=true
DB_HOST=localhost
REDIS_HOST=localhost
MINIO_ENDPOINT=localhost:9000
```

### 生产环境
```env
APP_ENV=production
APP_DEBUG=false
DB_HOST=postgres.production.com
REDIS_HOST=redis.production.com
MINIO_ENDPOINT=minio.production.com:9000
JWT_SECRET=your-production-secret-key
```

## 故障排查

### 配置文件未找到
- 确保配置文件在 `./config` 目录下
- 检查文件名是否正确（app.yaml, database.yaml）

### 环境变量未生效
- 确保 `.env` 文件在项目根目录
- 检查环境变量名是否正确
- 重启应用以加载新的环境变量

### 配置解析错误
- 检查 YAML 语法是否正确
- 确保缩进使用空格而非制表符
- 验证配置结构是否匹配 Go 结构体

## 相关文件

- `config/config.go` - 配置结构定义和加载逻辑
- `config/env.go` - 环境变量辅助函数
- `config/app.yaml` - 应用配置文件
- `config/database.yaml` - 数据库配置文件
- `.env.example` - 环境变量示例文件
