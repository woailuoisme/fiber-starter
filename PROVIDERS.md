# Application Providers

本文档详尽列出了项目中所有核心服务提供者（Providers）的契约方法及其功能说明，并附带了容器访问与 Facade 两种调用方式的示例。

## 目录

- [核心概念](#核心概念)
- [Auth (身份验证)](#auth-身份验证)
- [Cache (缓存)](#cache-缓存)
- [Config (配置)](#config-配置)
- [Database (数据库)](#database-数据库)
- [Hash (哈希)](#hash-哈希)
- [I18n (国际化)](#i18n-国际化)
- [Realtime (实时通信)](#realtime-实时通信)
- [Logging (日志)](#logging-日志)
- [Mail (邮件)](#mail-邮件)
- [Notification (通知)](#notification-通知)
- [Queue (队列)](#queue-队列)
- [RateLimiter (限流)](#ratelimiter-限流)
- [Schedule (任务调度)](#schedule-任务调度)
- [Search (搜索引擎)](#search-搜索引擎)
- [Storage (文件存储)](#storage-文件存储)
- [Validation (验证)](#validation-验证)

---

## 核心概念

### 全局容器访问 (Standard DI)

通过 `providers.App()` 获取全局 `Runtime` 实例。这种方式适合在结构体初始化或需要明确依赖关系的场景中使用。

```go
import "lfiber/app/Providers"

rt := providers.App()
db := rt.Connection
cfg := rt.Config
```

### Facade 访问 (Static Proxy)

每个 Provider 包都导出了直接调用的函数，这些函数会自动从全局容器中获取实例。这种方式代码更简洁，适合在控制器或辅助函数中快速使用。

```go
import "lfiber/app/Providers/Cache"

val, err := cache.Get("my-key")
```

---

## Auth (身份验证)

**契约**: `Guard` (`app/Providers/Auth/Contracts/guard.go`)

### Auth 方法列表

- `Check(c fiber.Ctx) bool`: 检查当前用户是否已通过身份验证。
- `Guest(c fiber.Ctx) bool`: 检查当前用户是否为游客（未登录）。
- `User(c fiber.Ctx) *models.User`: 获取当前已登录的用户模型。
- `UserIdentifier(c fiber.Ctx) string`: 获取当前用户的唯一标识符。
- `Id(c fiber.Ctx) int64`: 获取当前已登录用户的 ID。
- `SetUser(c fiber.Ctx, user *models.User)`: 手动设置当前上下文的用户。
- `Validate(credentials map[string]string) bool`: 仅校验凭据（如邮箱/密码）是否正确，不执行登录。
- `Attempt(c fiber.Ctx, credentials map[string]string) bool`: 尝试校验凭据并登录用户。
- `Login(c fiber.Ctx, user *models.User) error`: 将指定用户实例登录到应用。
- `LoginUsingId(c fiber.Ctx, id int64) error`: 根据用户 ID 登录用户。
- `Logout(c fiber.Ctx) error`: 注销当前用户并清除会话状态。

### Auth 使用示例

#### Auth 方案 A: 全局容器

```go
import "lfiber/app/Providers"

auth := providers.App().Auth
if auth.Check(c) {
    user := auth.User(c)
}
```

#### Auth 方案 B: Facade

```go
import "lfiber/app/Providers/Auth"

if auth.Check(c) {
    user := auth.User(c)
}
```

---

## Cache (缓存)

**契约**: `Store` (`app/Providers/Cache/Contracts/store.go`)

### Cache 方法列表

- `Get(key string) (string, error)`: 获取缓存字符串值。
- `GetBytes(key string) ([]byte, error)`: 获取原始字节数据。
- `GetJSON(key string, dest interface{}) error`: 获取 JSON 并反序列化到目标对象。
- `Set(key string, value interface{}, expiration time.Duration) error`: 存储缓存项。
- `Put(key string, value interface{}, expiration time.Duration) error`: `Set` 的别名。
- `Add(key string, value interface{}, expiration time.Duration) (bool, error)`: 仅在键不存在时存储。
- `Forever(key string, value interface{}) error`: 永久存储缓存。
- `Delete(key string) error`: 删除缓存项。
- `Forget(key string) error`: `Delete` 的别名。
- `DeletePattern(pattern string) error`: 根据 Glob 模式批量删除（如 `users:*`）。
- `Flush() error`: 清空所有缓存。
- `Exists(key string) (bool, error)`: 检查键是否存在。
- `Has(key string) (bool, error)`: `Exists` 的别名。
- `Pull(key string) (string, error)`: 获取并立即删除缓存。
- `Increment(key string) (int64, error)`: 原子递增。
- `Decrement(key string) (int64, error)`: 原子递减。
- `TTL(key string) (time.Duration, error)`: 获取剩余过期时间。
- `Expire(key string, expiration time.Duration) error`: 为现有键设置新过期时间。

### Cache 使用示例

#### Cache 方案 A: 全局容器

```go
import "lfiber/app/Providers"

providers.App().Cache.Set("key", "value", 10*time.Minute)
```

#### Cache 方案 B: Facade

```go
import "lfiber/app/Providers/Cache"

cache.Set("key", "value", 10*time.Minute)
val, _ := cache.Get("key")
```

---

## Config (配置)

**契约**: `Repository` (`app/Providers/Config/Contracts/repository.go`)

### Config 方法列表

- `Get(key string, defaultValue ...interface{}) interface{}`: 获取原始配置值。
- `Has(key string) bool`: 检查配置键是否存在。
- `All() map[string]interface{}`: 获取所有配置数据。
- `GetString(key string, defaultValue ...string) string`: 获取字符串配置。
- `GetInt(key string, defaultValue ...int) int`: 获取整数配置。
- `GetBool(key string, defaultValue ...bool) bool`: 获取布尔配置。
- `GetFloat64(key string, defaultValue ...float64) float64`: 获取浮点数配置。
- `Set(key string, value interface{})`: 运行时设置配置值。
- `Prepend(key string, value interface{})`: 向数组配置头部添加值。
- `Push(key string, value interface{})`: 向数组配置尾部添加值。

### Config 使用示例

#### Config 方案 A: 全局容器

```go
import "lfiber/app/Providers"

appName := providers.App().ConfigRepo.GetString("app.name")
```

#### Config 方案 B: Facade

```go
import "lfiber/app/Providers/Config"

appName := config.GetString("app.name")
```

---

## Database (数据库)

**契约**: `Manager` & `Connection` (`app/Providers/Database/Contracts/database.go`)

### Database Manager 方法列表 (管理多连接)

- `Connection(name ...string) Connection`: 根据名称获取连接（默认主连接）。
- `Reconnect(name ...string) (Connection, error)`: 重启指定连接。
- `Disconnect(name ...string) error`: 关闭指定连接。
- `Purge(name ...string)`: 关闭并从管理器中移除连接。
- `GetDefaultConnection() string`: 获取默认连接名。
- `SetDefaultConnection(name string)`: 设置默认连接名。
- `CloseAll() error`: 关闭所有连接。

### Database Connection 方法列表 (单一连接操作)

- `GetDB() (*sql.DB, error)`: 获取原生 `*sql.DB` 实例。
- `BunDB() *bun.DB`: 获取 Bun ORM 实例。
- `Dialect() (string, error)`: 获取数据库方言（如 `postgres`, `mysql`）。
- `GetName() string`: 获取连接定义的名称。
- `GetDriverName() string`: 获取驱动名称。
- `HealthCheck() error`: 检查连接健康状态。
- `GetStats() (map[string]interface{}, error)`: 获取连接统计信息。

### Database 使用示例

#### Database 方案 A: 全局容器

```go
import "lfiber/app/Providers"

db := providers.App().Connection.BunDB()
```

#### Database 方案 B: Facade

```go
import "lfiber/app/Providers/Database"

db := database.Connection().BunDB()
```

---

## Hash (哈希)

**契约**: `Hasher` (`app/Providers/Hash/Contracts/hasher.go`)

### Hash 方法列表

- `Make(value string) (string, error)`: 创建哈希值。
- `Check(value, hashedValue string) bool`: 校验明文与哈希是否匹配。
- `NeedsRehash(hashedValue string) bool`: 检查哈希是否需要重新加密（因配置变更）。
- `Info(hashedValue string) map[string]interface{}`: 获取哈希算法及参数信息。

### Hash 使用示例

#### Hash 方案 A: 全局容器

```go
import "lfiber/app/Providers"

hashed, _ := providers.App().Hash.Make("password")
```

#### Hash 方案 B: Facade

```go
import "lfiber/app/Providers/Hash"

hashed, _ := hash.Make("password")
isValid := hash.Check("password", hashed)
```

---

## I18n (国际化)

**契约**: `Translator` (`app/Providers/I18n/Contracts/translator.go`)

### I18n 方法列表

- `Trans(c fiber.Ctx, key string, params map[string]interface{}, locale ...string) string`: 翻译指定键。
- `Choice(c fiber.Ctx, key string, number int, params map[string]interface{}, locale ...string) string`: 复数形式翻译。
- `GetLocale(c fiber.Ctx) string`: 获取当前语言区域。
- `SetLocale(c fiber.Ctx, locale string) error`: 设置并持久化当前语言。
- `Middleware() fiber.Handler`: 获取自动处理语言切换的中间件。

### I18n 使用示例

#### I18n 方案 A: 全局容器

```go
import "lfiber/app/Providers"

msg := providers.App().Translator.Trans(c, "messages.welcome", nil)
```

#### I18n 方案 B: Facade

```go
import "lfiber/app/Providers/I18n"

msg := i18n.Trans(c, "messages.welcome", nil)
```

---

## Realtime (实时通信)

**契约**: `Manager` (`app/Providers/Realtime/Contracts/manager.go`)

### Realtime 方法列表

- `Handler() fiber.Handler`: 获取 WebSocket 主处理函数。
- `AuthHandler() fiber.Handler`: 获取用于 WebSocket 握手校验的中间件。
- `Dispatch(channel, event string, data any) error`: 向指定频道分发事件 and 数据。
- `Close() error`: 关闭实时通信服务。

### Realtime 使用示例

#### Realtime 方案 A: 全局容器

```go
import "lfiber/app/Providers"

providers.App().Realtime.Dispatch("public", "update", data)
```

#### Realtime 方案 B: Facade

```go
import "lfiber/app/Providers/Realtime"

realtime.Dispatch("public", "update", data)
```

---

## Logging (日志)

**契约**: `Logger` (`app/Providers/Logging/Contracts/logger.go`)

### Logging 方法列表

- `Channel(name string) Logger`: 切换到指定日志通道（如 `daily`, `slack`）。
- `Default() Logger`: 获取默认通道的 Logger。
- `Debug(msg string, fields ...zap.Field)`: 记录调试信息。
- `Info(msg string, fields ...zap.Field)`: 记录常规信息。
- `Warn(msg string, fields ...zap.Field)`: 记录警告信息。
- `Error(msg string, fields ...zap.Field)`: 记录错误信息。
- `Fatal(msg string, fields ...zap.Field)`: 记录致命错误并终止进程。
- `Panic(msg string, fields ...zap.Field)`: 记录错误并触发 Panic。
- `With(fields ...zap.Field) Logger`: 为后续日志附加固定字段。
- `GetZapLogger() *zap.Logger`: 获取底层 Zap 实例。

### Logging 使用示例

#### Logging 方案 A: 全局容器

```go
import "lfiber/app/Providers"

providers.App().Log.Info("User action", zap.String("key", "val"))
```

#### Logging 方案 B: Facade

```go
import "lfiber/app/Providers/Logging"

logging.Info("User action", zap.String("key", "val"))
logging.Channel("slack").Error("Critical error")
```

---

## Mail (邮件)

**契约**: `Mailer` & `Message` (`app/Providers/Mail/Contracts/mailer.go`)

### Mail 方法列表

- `To(to ...string) Message`: 初始化邮件并设置收件人。
- `Send(message Message) error`: 发送邮件对象。
- `Raw(to, subject, body string) error`: 快速发送文本/HTML 邮件。
- `Close() error`: 关闭邮件传输连接。

### Mail 使用示例

#### Mail 方案 A: 全局容器

```go
import "lfiber/app/Providers"

msg := providers.App().EmailService.To("user@test.com").Subject("Hi").Plain("Body")
providers.App().EmailService.Send(msg)
```

#### Mail 方案 B: Facade

```go
import "lfiber/app/Providers/Mail"

msg := mail.To("user@test.com").Subject("Hi").Plain("Body")
mail.Send(msg)
```

---

## Notification (通知)

**契约**: `Dispatcher` (`app/Providers/Notification/Contracts/notification.go`)

### Notification 方法列表

- `Send(notifiables interface{}, notification Notification) error`: 向目标发送通知。

### Notification 使用示例

#### Notification 方案 A: 全局容器

```go
import "lfiber/app/Providers"

providers.App().Notification.Send(user, &WelcomeNotification{})
```

#### Notification 方案 B: Facade

```go
import "lfiber/app/Providers/Notification"

notification.Send(user, &WelcomeNotification{})
```

---

## Queue (队列)

**契约**: `Manager` & `Queue` (`app/Providers/Queue/Contracts/queue.go`)

### Queue 方法列表

- `Drive(name ...string) Queue`: 获取指定驱动的队列。
- `Push(job Job) error`: 任务入队。
- `Later(delay time.Duration, job Job) error`: 延迟入队。
- `Register(job Job)`: 注册 Job 类型以便 Worker 反序列化。
- `StartWorker(queue ...string) error`: 启动 Worker。
- `InspectQueues() ([]QueueStatus, error)`: 检查队列状态。

### Queue 使用示例

#### Queue 方案 A: 全局容器

```go
import "lfiber/app/Providers"

providers.App().QueueService.Push(&MyJob{})
```

#### Queue 方案 B: Facade

```go
import "lfiber/app/Providers/Queue"

queue.Push(&MyJob{})
```

---

## RateLimiter (限流)

**契约**: `Limiter` (`app/Providers/RateLimiter/Contracts/limiter.go`)

### RateLimiter 方法列表

- `Strategy(name string) (configs.RateLimitConfig, bool)`: 获取指定名称的限流配置。
- `Middleware(name string) fiber.Handler`: 获取对应的限流中间件。

### RateLimiter 使用示例

#### RateLimiter 方案 A: 全局容器

```go
import "lfiber/app/Providers"

app.Use(providers.App().RateLimiter.Middleware("api"))
```

#### RateLimiter 方案 B: Facade

```go
import "lfiber/app/Providers/RateLimiter"

app.Use(ratelimiter.Middleware("api"))
```

---

## Schedule (任务调度)

**契约**: `Manager`, `Scheduler` & `Event` (`app/Providers/Schedule/Contracts/scheduler.go`)

### Schedule 方法列表

- `Job(job Contracts.Job) *Event`: 调度 Job。
- `Call(fn func() error) *Event`: 调度函数。
- `Command(command string, args ...string) *Event`: 调度 CLI 命令。
- `Run() error`: 启动调度器。

### Schedule 使用示例

#### Schedule 方案 A: 全局容器

```go
import "lfiber/app/Providers"

providers.App().ScheduleService.Command("db:seed").Daily()
```

#### Schedule 方案 B: Facade

```go
import "lfiber/app/Providers/Schedule"

schedule.Command("db:seed").Daily()
```

---

## Search (搜索引擎)

**契约**: `Engine` (`app/Providers/Search/Contracts/engine.go`)

### Search 方法列表

- `CreateIndex(uid string, primaryKey string) (*TaskInfo, error)`: 创建索引。
- `AddDocuments(indexUID string, documents interface{}) (*TaskInfo, error)`: 添加文档。
- `Search(indexUID string, query string, request *SearchRequest) (*SearchResponse, error)`: 搜索。

### Search 使用示例

#### Search 方案 A: 全局容器

```go
import "lfiber/app/Providers"

res, _ := providers.App().SearchService.Search("index", "query", nil)
```

#### Search 方案 B: Facade

```go
import "lfiber/app/Providers/Search"

res, _ := search.Search("index", "query", nil)
```

---

## Storage (文件存储)

**契约**: `StorageManager` & `Disk` (`app/Providers/Storage/Contracts/disk.go`)

### Storage 方法列表

- `Disk(name ...string) Disk`: 获取指定磁盘。
- `Get(path string) ([]byte, error)`: 读取文件。
- `Put(path string, contents []byte, options ...interface{}) error`: 写入文件。
- `Exists(path string) (bool, error)`: 检查是否存在。
- `Url(path string) string`: 获取公开 URL。

### Storage 使用示例

#### Storage 方案 A: 全局容器

```go
import "lfiber/app/Providers"

providers.App().Storage.Disk("s3").Put("file.txt", []byte("data"))
```

#### Storage 方案 B: Facade

```go
import "lfiber/app/Providers/Storage"

storage.Disk("s3").Put("file.txt", []byte("data"))
val, _ := storage.Get("file.txt")
```

---

## Validation (验证)

**契约**: `Factory`, `Validator` & `MessageBag` (`app/Providers/Validation/Contracts/factory.go` & `validator.go`)

### Validation 方法列表

- `Make(data any, rules map[string]string, messages map[string]string, attributes map[string]string) Validator`: 创建验证器。
- `Extend(rule string, extension validator.Func, message string) error`: 扩展规则。
- `Validate() error`: 执行验证。
- `Fails() bool`: 检查是否校验失败。
- `Errors() MessageBag`: 获取错误信息。

### Validation 使用示例

#### Validation 方案 A: 全局容器

```go
import "lfiber/app/Providers"

val := providers.App().Validation.Make(data, rules, nil, nil)
```

#### Validation 方案 B: Facade

```go
import "lfiber/app/Providers/Validation"

val := validation.Make(data, rules, nil, nil)
if val.Fails() {
    // ...
}
```
