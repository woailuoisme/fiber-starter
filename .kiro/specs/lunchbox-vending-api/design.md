# Design Document - 饭盒售货机后端 API 系统

## Overview

本设计文档描述了饭盒售货机后端 RESTful API 系统的技术架构和实现方案。该系统采用 Golang Fiber 框架开发，完全仿照 Laravel 框架的设计理念，为 Flutter 移动应用提供完整的后端支持。

系统核心特点：
- **Laravel 风格架构**：MVC + Service Layer 模式，依赖注入容器
- **高性能**：基于 Fiber 框架，支持高并发请求
- **完整功能**：认证、授权、缓存、队列、调度、WebSocket、媒体管理
- **双库存系统**：虚拟库存（下单）+ 实际库存（取货）
- **53 货道管理**：每台设备 53 个货道，精确库存控制
- **实时通信**：WebSocket 推送订单状态和支付通知

核心技术栈：
- **数据库**：PostgreSQL（主数据库）
- **缓存**：Redis（缓存和队列后端）
- **对象存储**：MinIO（媒体文件存储）
- **Web 框架**：Fiber v2
- **ORM**：GORM
- **队列**：Asynq
- **调度**：gocron
- **WebSocket**：gorilla/websocket
- **日志**：Zap

## Architecture

### 整体架构

系统采用分层架构设计，从上到下分为：

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                          │
│              (Flutter Mobile App / Admin)                │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                   API Gateway Layer                      │
│         (Fiber Router + Middleware + WebSocket)          │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                  Controller Layer                        │
│           (HTTP Request Handling + Validation)           │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                   Service Layer                          │
│              (Business Logic + Transactions)             │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                    Model Layer                           │
│              (GORM Models + Relationships)               │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                    │
│  (PostgreSQL + Redis + MinIO + Queue + WebSocket)       │
└─────────────────────────────────────────────────────────┘
```

### 目录结构（Laravel 风格）

```
.
├── app/
│   ├── controllers/          # 控制器层
│   │   ├── auth_controller.go
│   │   ├── city_controller.go
│   │   ├── device_controller.go
│   │   ├── product_controller.go
│   │   ├── order_controller.go
│   │   ├── payment_controller.go
│   │   └── media_controller.go
│   ├── models/              # 数据模型
│   │   ├── user.go
│   │   ├── city.go
│   │   ├── device.go
│   │   ├── device_channel.go
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── order_item.go
│   │   ├── payment.go
│   │   └── media.go
│   ├── services/            # 业务逻辑层
│   │   ├── auth_service.go
│   │   ├── city_service.go
│   │   ├── device_service.go
│   │   ├── product_service.go
│   │   ├── order_service.go
│   │   ├── payment_service.go
│   │   ├── media_service.go
│   │   ├── cache_service.go
│   │   └── websocket_service.go
│   ├── middleware/          # 中间件
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   ├── rate_limiter.go
│   │   └── error_handler.go
│   ├── providers/           # 服务提供者（DI 容器）
│   │   ├── app_provider.go
│   │   ├── database_provider.go
│   │   ├── cache_provider.go
│   │   └── queue_provider.go
│   ├── requests/            # 请求验证
│   │   ├── auth_request.go
│   │   ├── order_request.go
│   │   └── product_request.go
│   ├── responses/           # 响应格式化
│   │   ├── api_response.go
│   │   ├── paginated_response.go
│   │   └── error_response.go
│   ├── exceptions/          # 自定义异常
│   │   ├── api_exception.go
│   │   ├── validation_exception.go
│   │   ├── authentication_exception.go
│   │   └── not_found_exception.go
│   ├── jobs/                # 队列任务
│   │   ├── send_email_job.go
│   │   ├── process_payment_job.go
│   │   └── cleanup_orders_job.go
│   ├── helpers/             # 辅助函数
│   │   ├── jwt_helper.go
│   │   ├── hash_helper.go
│   │   └── qrcode_helper.go
│   └── enums/               # 枚举类型
│       ├── order_status.go
│       ├── payment_status.go
│       └── device_status.go
├── config/                   # 配置文件
│   ├── app.yaml
│   ├── database.yaml
│   └── config.go
├── database/                 # 数据库相关
│   ├── migrations/          # 迁移文件
│   │   ├── 001_create_users_table.go
│   │   ├── 002_create_cities_table.go
│   │   ├── 003_create_devices_table.go
│   │   ├── 004_create_device_channels_table.go
│   │   ├── 005_create_products_table.go
│   │   ├── 006_create_orders_table.go
│   │   ├── 007_create_payments_table.go
│   │   └── 008_create_media_table.go
│   ├── seeders/             # 数据填充
│   │   ├── user_seeder.go
│   │   ├── city_seeder.go
│   │   └── product_seeder.go
│   └── connection.go        # 数据库连接
├── routes/                   # 路由定义
│   ├── api.go               # API 路由
│   └── websocket.go         # WebSocket 路由
├── storage/                  # 存储目录
│   ├── logs/                # 日志文件
│   ├── media/               # 媒体文件
│   └── cache/               # 缓存文件
├── command/                  # 命令行工具
│   ├── root.go
│   ├── migrate.go
│   ├── seed.go
│   ├── schedule.go
│   ├── queue.go
│   └── websocket.go
├── tests/                    # 测试文件
│   ├── unit/
│   └── integration/
└── main.go                  # 应用入口
```

## Components and Interfaces

### 1. 控制器层（Controllers）

控制器负责处理 HTTP 请求，验证输入，调用服务层，返回响应。

**接口设计**：
```go
type Controller interface {
    Index(c *fiber.Ctx) error    // GET /resource
    Show(c *fiber.Ctx) error     // GET /resource/:id
    Store(c *fiber.Ctx) error    // POST /resource
    Update(c *fiber.Ctx) error   // PUT /resource/:id
    Destroy(c *fiber.Ctx) error  // DELETE /resource/:id
}
```

**示例：OrderController**
```go
type OrderController struct {
    orderService *services.OrderService
}

func (ctrl *OrderController) Store(c *fiber.Ctx) error {
    // 1. 验证请求
    req := new(requests.CreateOrderRequest)
    if err := c.BodyParser(req); err != nil {
        return exceptions.NewValidationException("Invalid request body")
    }
    
    // 2. 获取当前用户
    userID := c.Locals("user_id").(uint)
    
    // 3. 调用服务层
    order, err := ctrl.orderService.CreateOrder(userID, req)
    if err != nil {
        return err
    }
    
    // 4. 返回响应
    return responses.Success(c, "Order created successfully", order)
}
```

### 2. 服务层（Services）

服务层封装业务逻辑，处理事务，调用模型层。

**接口设计**：
```go
type OrderService interface {
    CreateOrder(userID uint, req *requests.CreateOrderRequest) (*models.Order, error)
    GetOrder(orderID uint) (*models.Order, error)
    GetUserOrders(userID uint, page int) (*responses.PaginatedResponse, error)
    CancelOrder(orderID uint) error
    VerifyQRCode(qrCode string) (*models.Order, error)
    CompleteOrder(orderID uint) error
}
```

**关键业务逻辑**：
- 订单创建：验证库存 → 扣减虚拟库存 → 生成二维码 → 创建订单
- 订单支付：验证订单状态 → 调用支付接口 → 更新订单状态
- 订单取货：验证二维码 → 检查订单状态 → 返回货道信息 → 扣减实际库存

### 3. 模型层（Models）

模型层定义数据结构和关系，使用 GORM ORM。

**核心模型关系**：
```go
// User 用户模型
type User struct {
    gorm.Model
    Name     string
    Email    string `gorm:"uniqueIndex"`
    Password string
    Role     string // user, admin
    Orders   []Order
}

// City 城市模型
type City struct {
    gorm.Model
    Name    string
    Status  string // active, inactive
    Devices []Device
}

// Device 设备模型
type Device struct {
    gorm.Model
    CityID    uint
    City      City
    Name      string
    Code      string `gorm:"uniqueIndex"`
    Latitude  float64
    Longitude float64
    Status    string // online, offline, fault
    Channels  []DeviceChannel
    Orders    []Order
}

// DeviceChannel 货道模型
type DeviceChannel struct {
    gorm.Model
    DeviceID      uint
    Device        Device
    ChannelNumber int `gorm:"check:channel_number >= 1 AND channel_number <= 53"`
    ProductID     *uint
    Product       *Product
    VirtualStock  int `gorm:"check:virtual_stock >= 0 AND virtual_stock <= 4"`
    ActualStock   int `gorm:"check:actual_stock >= 0 AND actual_stock <= 4"`
    MaxCapacity   int `gorm:"default:4"`
    Status        string // normal, fault, maintenance, disabled
}

// Product 产品模型
type Product struct {
    gorm.Model
    Name        string
    Description string
    Price       int // 以分为单位
    CategoryID  uint
    Category    Category
    ImageURL    string
}

// Order 订单模型
type Order struct {
    gorm.Model
    UserID      uint
    User        User
    DeviceID    uint
    Device      Device
    OrderNumber string `gorm:"uniqueIndex"`
    TotalAmount int
    Status      string // pending, paid, completed, cancelled, timeout
    QRCode      string `gorm:"uniqueIndex"`
    PaidAt      *time.Time
    CompletedAt *time.Time
    Items       []OrderItem
    Payment     *Payment
}

// OrderItem 订单项模型
type OrderItem struct {
    gorm.Model
    OrderID       uint
    Order         Order
    ChannelID     uint
    Channel       DeviceChannel
    ProductID     uint
    Product       Product
    Quantity      int
    Price         int
    ChannelNumber int
}

// Payment 支付模型
type Payment struct {
    gorm.Model
    OrderID         uint
    Order           Order
    PaymentMethod   string // wechat, alipay
    PaymentNumber   string `gorm:"uniqueIndex"`
    Amount          int
    Status          string // pending, success, failed, refunded
    TransactionID   string
    PaidAt          *time.Time
}

// Media 媒体模型
type Media struct {
    gorm.Model
    ModelType        string // Product, User, etc.
    ModelID          uint
    CollectionName   string // images, avatars, documents
    Name             string
    FileName         string
    MimeType         string
    Disk             string // local, minio, s3
    Size             int64
    Manipulations    datatypes.JSON
    CustomProperties datatypes.JSON
    OrderColumn      int
}
```

### 4. 中间件（Middleware）

**认证中间件**：
```go
func AuthMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        if token == "" {
            return exceptions.NewAuthenticationException("Missing token")
        }
        
        claims, err := helpers.VerifyJWT(token)
        if err != nil {
            return exceptions.NewAuthenticationException("Invalid token")
        }
        
        c.Locals("user_id", claims.UserID)
        c.Locals("user_role", claims.Role)
        return c.Next()
    }
}
```

**错误处理中间件**：
```go
func ErrorHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        err := c.Next()
        if err == nil {
            return nil
        }
        
        // 处理 ApiException
        if apiErr, ok := err.(*exceptions.ApiException); ok {
            return responses.Error(c, apiErr.Code, apiErr.Message, apiErr.Errors)
        }
        
        // 处理其他错误
        return responses.Error(c, 500, "Internal server error", nil)
    }
}
```

### 5. 响应格式化（Responses）

**统一响应结构**：
```go
type ApiResponse struct {
    Success bool        `json:"success"`
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Errors  interface{} `json:"errors,omitempty"`
    Debugger *Debugger  `json:"debugger,omitempty"`
}

type Debugger struct {
    Exception   string   `json:"exception"`
    File        string   `json:"file"`
    Line        int      `json:"line"`
    Trace       []string `json:"trace"`
    RequestTime float64  `json:"request_time"`
    MemoryUsage string   `json:"memory_usage"`
    QueryCount  int      `json:"query_count"`
}

type PaginatedResponse struct {
    Items []interface{} `json:"items"`
    Meta  PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
    CurrentPage int  `json:"current_page"`
    PerPage     int  `json:"per_page"`
    LastPage    int  `json:"last_page"`
    HasMore     bool `json:"has_more"`
    Total       int64 `json:"total"`
    From        int  `json:"from"`
    To          int  `json:"to"`
}
```

### 6. 自定义异常（Exceptions）

```go
type ApiException struct {
    Message string
    Code    int
    Errors  map[string][]string
}

func NewValidationException(message string) *ApiException {
    return &ApiException{Message: message, Code: 422}
}

func NewAuthenticationException(message string) *ApiException {
    return &ApiException{Message: message, Code: 401}
}

func NewAuthorizationException(message string) *ApiException {
    return &ApiException{Message: message, Code: 403}
}

func NewNotFoundException(message string) *ApiException {
    return &ApiException{Message: message, Code: 404}
}
```

## Data Models

### 数据库表结构

**users 表**：
```sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role ENUM('user', 'admin') DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_email (email),
    INDEX idx_role (role)
);
```

**cities 表**：
```sql
CREATE TABLE cities (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status ENUM('active', 'inactive') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_status (status)
);
```

**devices 表**：
```sql
CREATE TABLE devices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    city_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    status ENUM('online', 'offline', 'fault') DEFAULT 'offline',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (city_id) REFERENCES cities(id),
    INDEX idx_city_id (city_id),
    INDEX idx_code (code),
    INDEX idx_status (status)
);
```

**device_channels 表**：
```sql
CREATE TABLE device_channels (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    device_id BIGINT UNSIGNED NOT NULL,
    channel_number INT NOT NULL CHECK (channel_number >= 1 AND channel_number <= 53),
    product_id BIGINT UNSIGNED,
    virtual_stock INT DEFAULT 0 CHECK (virtual_stock >= 0 AND virtual_stock <= 4),
    actual_stock INT DEFAULT 0 CHECK (actual_stock >= 0 AND actual_stock <= 4),
    max_capacity INT DEFAULT 4,
    status ENUM('normal', 'fault', 'maintenance', 'disabled') DEFAULT 'normal',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (device_id) REFERENCES devices(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    UNIQUE KEY uk_device_channel (device_id, channel_number),
    INDEX idx_device_id (device_id),
    INDEX idx_product_id (product_id),
    INDEX idx_status (status)
);
```

**products 表**：
```sql
CREATE TABLE products (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INT NOT NULL,
    category_id BIGINT UNSIGNED,
    image_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_category_id (category_id),
    INDEX idx_price (price)
);
```

**orders 表**：
```sql
CREATE TABLE orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    device_id BIGINT UNSIGNED NOT NULL,
    order_number VARCHAR(100) UNIQUE NOT NULL,
    total_amount INT NOT NULL,
    status ENUM('pending', 'paid', 'completed', 'cancelled', 'timeout') DEFAULT 'pending',
    qr_code VARCHAR(255) UNIQUE NOT NULL,
    paid_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (device_id) REFERENCES devices(id),
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_order_number (order_number),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

**order_items 表**：
```sql
CREATE TABLE order_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    channel_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    quantity INT NOT NULL,
    price INT NOT NULL,
    channel_number INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (channel_id) REFERENCES device_channels(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    INDEX idx_order_id (order_id)
);
```

**payments 表**：
```sql
CREATE TABLE payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    payment_method ENUM('wechat', 'alipay') NOT NULL,
    payment_number VARCHAR(100) UNIQUE NOT NULL,
    amount INT NOT NULL,
    status ENUM('pending', 'success', 'failed', 'refunded') DEFAULT 'pending',
    transaction_id VARCHAR(255),
    paid_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    INDEX idx_order_id (order_id),
    INDEX idx_payment_number (payment_number),
    INDEX idx_status (status)
);
```

**media 表**：
```sql
CREATE TABLE media (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    model_type VARCHAR(255) NOT NULL,
    model_id BIGINT UNSIGNED NOT NULL,
    collection_name VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(255),
    disk VARCHAR(50) DEFAULT 'local',
    size BIGINT,
    manipulations JSON,
    custom_properties JSON,
    order_column INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_model (model_type, model_id),
    INDEX idx_collection (collection_name),
    INDEX idx_order (order_column)
);
```

## Error Handling

### 错误处理策略

1. **全局错误处理中间件**：捕获所有未处理的错误
2. **自定义异常类型**：不同类型的错误使用不同的异常类
3. **统一错误响应**：所有错误返回统一的 JSON 格式
4. **调试模式**：开发环境返回详细的错误信息和堆栈跟踪

### 错误响应示例

**生产环境**：
```json
{
  "success": false,
  "code": 422,
  "message": "Validation failed",
  "errors": {
    "email": ["The email field is required."],
    "password": ["The password must be at least 8 characters."]
  }
}
```

**开发环境**：
```json
{
  "success": false,
  "code": 500,
  "message": "Database connection failed",
  "errors": null,
  "debugger": {
    "exception": "DatabaseException",
    "file": "/app/database/connection.go",
    "line": 45,
    "trace": [
      "database.Connect() at connection.go:45",
      "main.main() at main.go:20"
    ],
    "request_time": 125.5,
    "memory_usage": "2.5MB",
    "query_count": 3
  }
}
```

## Testing Strategy

### 测试方法

1. **单元测试**：测试单个函数和方法
2. **集成测试**：测试多个组件的集成
3. **API 测试**：测试 HTTP 端点
4. **性能测试**：测试系统性能和并发能力

### 测试工具

- **testing**：Go 标准测试库
- **testify**：断言和 Mock 库
- **httptest**：HTTP 测试工具
- **gomock**：Mock 生成工具

### 测试覆盖率目标

- 核心业务逻辑：> 80%
- 服务层：> 70%
- 控制器层：> 60%
- 整体覆盖率：> 65%

## 继续设计文档...

由于文档较长，我将继续添加剩余部分。


## 核心业务流程设计

### 1. 用户下单流程

```
用户选择产品 → 验证库存 → 创建订单 → 扣减虚拟库存 → 生成二维码 → 返回订单信息
```

**详细步骤**：
1. 用户在 Flutter App 选择设备和产品（可多选）
2. 前端调用 `POST /api/orders` 创建订单
3. 后端验证：
   - 用户是否登录
   - 设备是否在线
   - 货道是否正常
   - 虚拟库存是否充足
4. 开启数据库事务
5. 扣减各货道的虚拟库存
6. 创建订单记录（status=pending）
7. 创建订单项记录
8. 生成唯一的取货二维码
9. 提交事务
10. 返回订单信息和支付参数

### 2. 支付流程

```
用户发起支付 → 调用支付接口 → 等待回调 → 更新订单状态 → 推送通知
```

**详细步骤**：
1. 用户选择支付方式（微信/支付宝）
2. 前端调用 `POST /api/orders/:id/pay`
3. 后端生成支付参数（二维码、订单号等）
4. 返回支付参数给前端
5. 用户完成支付
6. 支付平台回调 `POST /api/payments/callback`
7. 验证回调签名
8. 更新订单状态为 paid
9. 记录支付时间
10. 通过 WebSocket 推送支付成功通知

### 3. 取货流程

```
用户到设备 → 扫描二维码 → 验证订单 → 返回货道信息 → 设备出货 → 扣减实际库存
```

**详细步骤**：
1. 用户到达设备，打开取货二维码
2. 设备扫描二维码
3. 设备调用 `POST /api/orders/verify-qrcode`
4. 后端验证：
   - 二维码是否有效
   - 订单是否已支付
   - 订单是否已取货
   - 订单是否超时
5. 返回货道编号列表
6. 设备按顺序出货
7. 每个货道出货成功后，设备调用 `POST /api/orders/:id/complete-item`
8. 后端扣减对应货道的实际库存
9. 所有货道出货完成后，更新订单状态为 completed
10. 通过 WebSocket 推送取货成功通知

### 4. 订单超时处理

```
定时任务扫描 → 检查超时订单 → 取消订单 → 恢复虚拟库存
```

**详细步骤**：
1. 调度器每分钟执行一次检查
2. 查询所有 status=pending 且创建时间超过 30 分钟的订单
3. 查询所有 status=paid 且支付时间超过 24 小时的订单
4. 对每个超时订单：
   - 更新订单状态为 timeout
   - 恢复各货道的虚拟库存
   - 记录日志
5. 如果订单已支付，触发退款流程

## 缓存策略

### 缓存层级

1. **应用层缓存**：使用 Redis 缓存热点数据
2. **数据库查询缓存**：GORM 查询结果缓存
3. **HTTP 响应缓存**：静态资源和不常变化的 API 响应

### 缓存键设计

```
# 城市列表
cache:cities:list

# 城市详情
cache:city:{id}

# 设备列表（按城市）
cache:devices:city:{city_id}

# 设备详情
cache:device:{id}

# 设备货道列表
cache:device:{id}:channels

# 产品详情
cache:product:{id}

# 用户信息
cache:user:{id}
```

### 缓存失效策略

1. **主动失效**：数据更新时立即清除相关缓存
2. **被动失效**：设置 TTL，过期自动清除
3. **缓存预热**：系统启动时预加载热点数据

### 缓存实现

```go
type CacheService struct {
    redis *redis.Client
}

func (s *CacheService) Get(key string, dest interface{}) error {
    val, err := s.redis.Get(context.Background(), key).Result()
    if err == redis.Nil {
        return ErrCacheNotFound
    }
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (s *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return s.redis.Set(context.Background(), key, data, ttl).Err()
}

func (s *CacheService) Delete(key string) error {
    return s.redis.Del(context.Background(), key).Err()
}

func (s *CacheService) DeletePattern(pattern string) error {
    keys, err := s.redis.Keys(context.Background(), pattern).Result()
    if err != nil {
        return err
    }
    if len(keys) > 0 {
        return s.redis.Del(context.Background(), keys...).Err()
    }
    return nil
}
```

## 队列系统设计

### 队列任务类型

1. **邮件发送任务**：SendEmailJob
2. **支付处理任务**：ProcessPaymentJob
3. **订单清理任务**：CleanupOrdersJob
4. **库存同步任务**：SyncInventoryJob
5. **通知推送任务**：PushNotificationJob

### 队列实现（使用 Asynq）

```go
// 定义任务
type SendEmailJob struct {
    To      string
    Subject string
    Body    string
}

func (j *SendEmailJob) ProcessTask(ctx context.Context, task *asynq.Task) error {
    var payload SendEmailJob
    if err := json.Unmarshal(task.Payload(), &payload); err != nil {
        return err
    }
    
    // 发送邮件逻辑
    return sendEmail(payload.To, payload.Subject, payload.Body)
}

// 分发任务
func DispatchSendEmail(to, subject, body string) error {
    payload, err := json.Marshal(SendEmailJob{
        To:      to,
        Subject: subject,
        Body:    body,
    })
    if err != nil {
        return err
    }
    
    task := asynq.NewTask("email:send", payload)
    _, err = client.Enqueue(task, asynq.Queue("default"))
    return err
}

// 队列工作进程
func StartQueueWorker() {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: "localhost:6379"},
        asynq.Config{
            Concurrency: 10,
            Queues: map[string]int{
                "critical": 6,
                "default":  3,
                "low":      1,
            },
        },
    )
    
    mux := asynq.NewServeMux()
    mux.HandleFunc("email:send", new(SendEmailJob).ProcessTask)
    mux.HandleFunc("payment:process", new(ProcessPaymentJob).ProcessTask)
    
    if err := srv.Run(mux); err != nil {
        log.Fatal(err)
    }
}
```

## 任务调度系统设计

### 调度任务列表

1. **清理过期订单**：每分钟执行
2. **同步库存数据**：每 5 分钟执行
3. **生成日报**：每天凌晨 1 点执行
4. **清理临时文件**：每天凌晨 2 点执行
5. **数据备份**：每天凌晨 3 点执行

### 调度实现（使用 gocron）

```go
func SetupScheduler() {
    s := gocron.NewScheduler(time.UTC)
    
    // 每分钟清理过期订单
    s.Every(1).Minute().Do(func() {
        log.Println("Cleaning up expired orders...")
        CleanupExpiredOrders()
    })
    
    // 每 5 分钟同步库存
    s.Every(5).Minutes().Do(func() {
        log.Println("Syncing inventory...")
        SyncInventory()
    })
    
    // 每天凌晨 1 点生成日报
    s.Every(1).Day().At("01:00").Do(func() {
        log.Println("Generating daily report...")
        GenerateDailyReport()
    })
    
    // 每天凌晨 2 点清理临时文件
    s.Every(1).Day().At("02:00").Do(func() {
        log.Println("Cleaning up temp files...")
        CleanupTempFiles()
    })
    
    s.StartAsync()
}

func CleanupExpiredOrders() {
    // 查找超时订单
    var orders []models.Order
    db.Where("status = ? AND created_at < ?", "pending", time.Now().Add(-30*time.Minute)).Find(&orders)
    
    for _, order := range orders {
        // 开启事务
        tx := db.Begin()
        
        // 更新订单状态
        tx.Model(&order).Update("status", "timeout")
        
        // 恢复虚拟库存
        for _, item := range order.Items {
            tx.Model(&models.DeviceChannel{}).
                Where("id = ?", item.ChannelID).
                UpdateColumn("virtual_stock", gorm.Expr("virtual_stock + ?", item.Quantity))
        }
        
        tx.Commit()
    }
}
```

## WebSocket 实时通信设计

### WebSocket 架构

```
Client (Flutter App)
    ↓ (WebSocket Connection)
WebSocket Server (Fiber + gorilla/websocket)
    ↓
Channel Manager (管理连接和频道)
    ↓
Event Broadcaster (广播事件)
```

### 频道设计

1. **用户私有频道**：`user.{user_id}` - 接收个人通知
2. **订单频道**：`order.{order_id}` - 接收订单状态更新
3. **设备频道**：`device.{device_id}` - 接收设备状态更新
4. **公共频道**：`public` - 接收系统公告

### WebSocket 实现

```go
type WebSocketManager struct {
    clients   map[*websocket.Conn]*Client
    broadcast chan *Message
    register  chan *Client
    unregister chan *Client
    mu        sync.RWMutex
}

type Client struct {
    conn     *websocket.Conn
    userID   uint
    channels []string
    send     chan []byte
}

type Message struct {
    Channel string      `json:"channel"`
    Event   string      `json:"event"`
    Data    interface{} `json:"data"`
}

func (m *WebSocketManager) Run() {
    for {
        select {
        case client := <-m.register:
            m.mu.Lock()
            m.clients[client.conn] = client
            m.mu.Unlock()
            
        case client := <-m.unregister:
            m.mu.Lock()
            if _, ok := m.clients[client.conn]; ok {
                delete(m.clients, client.conn)
                close(client.send)
            }
            m.mu.Unlock()
            
        case message := <-m.broadcast:
            m.mu.RLock()
            for _, client := range m.clients {
                if m.isSubscribed(client, message.Channel) {
                    select {
                    case client.send <- m.encodeMessage(message):
                    default:
                        close(client.send)
                        delete(m.clients, client.conn)
                    }
                }
            }
            m.mu.RUnlock()
        }
    }
}

func (m *WebSocketManager) Broadcast(channel, event string, data interface{}) {
    m.broadcast <- &Message{
        Channel: channel,
        Event:   event,
        Data:    data,
    }
}

// 使用示例
func NotifyOrderPaid(orderID uint, userID uint) {
    wsManager.Broadcast(
        fmt.Sprintf("user.%d", userID),
        "order.paid",
        map[string]interface{}{
            "order_id": orderID,
            "message":  "Payment successful",
        },
    )
}
```

### WebSocket 事件类型

1. **order.created** - 订单创建
2. **order.paid** - 订单支付成功
3. **order.completed** - 订单完成
4. **order.cancelled** - 订单取消
5. **device.online** - 设备上线
6. **device.offline** - 设备离线
7. **inventory.low** - 库存预警

## 媒体文件管理设计

### 媒体库架构

```
Upload Request
    ↓
MediaController
    ↓
MediaService (处理上传、生成缩略图)
    ↓
Storage Driver (Local/MinIO/S3)
    ↓
Media Model (保存元数据)
```

### 存储驱动接口

```go
type StorageDriver interface {
    Put(path string, file io.Reader) (string, error)
    Get(path string) (io.Reader, error)
    Delete(path string) error
    URL(path string) string
}

// 本地存储驱动
type LocalDriver struct {
    basePath string
}

func (d *LocalDriver) Put(path string, file io.Reader) (string, error) {
    fullPath := filepath.Join(d.basePath, path)
    os.MkdirAll(filepath.Dir(fullPath), 0755)
    
    out, err := os.Create(fullPath)
    if err != nil {
        return "", err
    }
    defer out.Close()
    
    _, err = io.Copy(out, file)
    return path, err
}

// MinIO 存储驱动
type MinIODriver struct {
    client *minio.Client
    bucket string
}

func (d *MinIODriver) Put(path string, file io.Reader) (string, error) {
    _, err := d.client.PutObject(
        context.Background(),
        d.bucket,
        path,
        file,
        -1,
        minio.PutObjectOptions{},
    )
    return path, err
}
```

### 图片处理

```go
func (s *MediaService) ProcessImage(file multipart.File, filename string) (*models.Media, error) {
    // 1. 读取原图
    img, err := imaging.Decode(file)
    if err != nil {
        return nil, err
    }
    
    // 2. 生成缩略图
    thumbnails := map[string]image.Image{
        "thumb":  imaging.Resize(img, 150, 0, imaging.Lanczos),
        "medium": imaging.Resize(img, 500, 0, imaging.Lanczos),
        "large":  imaging.Resize(img, 1200, 0, imaging.Lanczos),
    }
    
    // 3. 保存原图和缩略图
    paths := make(map[string]string)
    for size, thumbnail := range thumbnails {
        path := fmt.Sprintf("media/%s/%s", size, filename)
        buf := new(bytes.Buffer)
        imaging.Encode(buf, thumbnail, imaging.JPEG)
        
        url, err := s.storage.Put(path, buf)
        if err != nil {
            return nil, err
        }
        paths[size] = url
    }
    
    // 4. 创建 Media 记录
    media := &models.Media{
        FileName:      filename,
        MimeType:      "image/jpeg",
        Disk:          s.config.Storage.Driver,
        Manipulations: paths,
    }
    
    return media, s.db.Create(media).Error
}
```

## API 端点设计

### 认证相关

```
POST   /api/auth/register          # 用户注册
POST   /api/auth/login             # 用户登录
POST   /api/auth/refresh           # 刷新 Token
POST   /api/auth/logout            # 用户登出
GET    /api/auth/me                # 获取当前用户信息
```

### 城市管理

```
GET    /api/cities                 # 获取城市列表
GET    /api/cities/:id             # 获取城市详情
POST   /api/cities                 # 创建城市（管理员）
PUT    /api/cities/:id             # 更新城市（管理员）
DELETE /api/cities/:id             # 删除城市（管理员）
```

### 设备管理

```
GET    /api/devices                # 获取设备列表
GET    /api/devices/:id            # 获取设备详情
GET    /api/cities/:id/devices     # 获取城市下的设备
POST   /api/devices                # 创建设备（管理员）
PUT    /api/devices/:id            # 更新设备（管理员）
DELETE /api/devices/:id            # 删除设备（管理员）
POST   /api/devices/:id/location   # 上报设备位置
GET    /api/devices/:id/channels   # 获取设备货道列表
PUT    /api/devices/:id/channels/:channel_number  # 更新货道信息
```

### 产品管理

```
GET    /api/products               # 获取产品列表
GET    /api/products/:id           # 获取产品详情
GET    /api/devices/:id/products   # 获取设备的产品列表
POST   /api/products               # 创建产品（管理员）
PUT    /api/products/:id           # 更新产品（管理员）
DELETE /api/products/:id           # 删除产品（管理员）
```

### 订单管理

```
POST   /api/orders                 # 创建订单
GET    /api/orders                 # 获取订单列表
GET    /api/orders/:id             # 获取订单详情
POST   /api/orders/:id/pay         # 支付订单
POST   /api/orders/:id/cancel      # 取消订单
POST   /api/orders/verify-qrcode   # 验证取货二维码
POST   /api/orders/:id/complete-item  # 完成单个货道出货
```

### 支付管理

```
POST   /api/payments/callback      # 支付回调
GET    /api/payments/:id           # 获取支付详情
POST   /api/payments/:id/refund    # 申请退款
```

### 媒体管理

```
POST   /api/media/upload           # 上传文件
GET    /api/media/:id              # 获取媒体详情
DELETE /api/media/:id              # 删除媒体文件
```

### WebSocket

```
WS     /ws                         # WebSocket 连接端点
```

## 性能优化策略

### 1. 数据库优化

- 为常用查询字段添加索引
- 使用复合索引优化多条件查询
- 使用数据库连接池
- 避免 N+1 查询问题（使用 Preload）
- 使用批量插入和更新

### 2. 缓存优化

- 缓存热点数据（城市列表、设备列表）
- 使用 Redis 缓存查询结果
- 实现缓存预热机制
- 合理设置缓存过期时间

### 3. API 优化

- 使用 gzip 压缩响应
- 实现 API 响应缓存
- 使用分页减少数据传输
- 实现字段过滤（只返回需要的字段）

### 4. 并发优化

- 使用 Goroutine 处理异步任务
- 使用 Channel 进行并发通信
- 使用 sync.WaitGroup 等待并发任务完成
- 使用 Context 控制超时和取消

## 安全措施

### 1. 认证和授权

- JWT Token 认证
- Token 过期时间控制
- 刷新 Token 机制
- 基于角色的访问控制（RBAC）

### 2. 数据安全

- 密码使用 bcrypt 加密
- 敏感数据加密存储
- HTTPS 传输加密
- SQL 注入防护（ORM 参数化查询）

### 3. API 安全

- 请求频率限制
- CORS 跨域配置
- XSS 防护
- CSRF 防护
- 输入验证和过滤

### 4. 日志和监控

- 记录所有安全事件
- 记录登录失败尝试
- 记录权限拒绝事件
- 实时监控异常请求

## 部署架构

### 生产环境架构

```
                    ┌─────────────┐
                    │   Nginx     │
                    │  (反向代理)  │
                    └──────┬──────┘
                           │
              ┌────────────┴────────────┐
              │                         │
        ┌─────▼─────┐            ┌─────▼─────┐
        │  Fiber    │            │  Fiber    │
        │  App 1    │            │  App 2    │
        └─────┬─────┘            └─────┬─────┘
              │                         │
              └────────────┬────────────┘
                           │
              ┌────────────┴────────────┐
              │                         │
        ┌─────▼─────┐            ┌─────▼─────┐
        │PostgreSQL │            │   Redis   │
        │  (主从)    │            │  (集群)    │
        └─────┬─────┘            └─────┬─────┘
              │                         │
              └────────────┬────────────┘
                           │
                     ┌─────▼─────┐
                     │   MinIO   │
                     │ (对象存储) │
                     └───────────┘
```

### Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 3000
CMD ["./main"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - MINIO_ENDPOINT=minio:9000
    depends_on:
      - postgres
      - redis
      - minio
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: lunchbox_vending
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"
    restart: unless-stopped

  queue:
    build: .
    command: ["./main", "queue:work"]
    depends_on:
      - redis
      - postgres
    restart: unless-stopped

  scheduler:
    build: .
    command: ["./main", "schedule:run"]
    depends_on:
      - redis
      - postgres
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

## 监控和日志

### 日志系统

使用 Zap 结构化日志：

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("Order created",
    zap.Uint("order_id", order.ID),
    zap.Uint("user_id", order.UserID),
    zap.Int("total_amount", order.TotalAmount),
)

logger.Error("Payment failed",
    zap.Uint("order_id", order.ID),
    zap.Error(err),
)
```

### 监控指标

1. **系统指标**：CPU、内存、磁盘使用率
2. **应用指标**：请求数、响应时间、错误率
3. **业务指标**：订单数、支付成功率、库存预警
4. **数据库指标**：连接数、查询时间、慢查询

## 总结

本设计文档详细描述了饭盒售货机后端 API 系统的技术架构、核心组件、数据模型、业务流程和实现方案。系统采用 Laravel 风格的分层架构，使用 Golang 生态中最流行的开源包，确保了高性能、高可靠性和良好的可维护性。

核心特性包括：
- 完整的用户认证和授权系统
- 53 货道管理和双库存系统
- 订单创建、支付和取货完整流程
- WebSocket 实时通信
- 队列和调度系统
- 媒体文件管理
- 完善的缓存和性能优化策略
- 全面的安全措施

系统已准备好进入实现阶段。
