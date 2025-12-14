# Design Document - Go Framework 全面优化

## Overview

本设计文档描述了对饭盒售货机后端 API 系统进行全面优化的技术方案。优化目标是将系统升级到生产级别，使用 Go 1.24 和最新的生态系统包，完全参照 Laravel 12 框架的最佳实践，实现高性能、高可用、高安全性的企业级应用。

### 核心设计理念

1. **Laravel 风格架构**：完全参照 Laravel 12 的设计模式和最佳实践
2. **清晰的分层架构**：Controller → Service → Repository → Model
3. **依赖注入**：使用服务容器管理所有依赖
4. **接口驱动**：面向接口编程，提升可测试性
5. **配置驱动**：所有行为通过配置控制
6. **约定优于配置**：遵循 Go 和 Laravel 的约定

### 技术栈升级

**核心框架**：
- Go 1.24（最新稳定版）
- Fiber v3（高性能 Web 框架）
- GORM v2（ORM 框架）
- Wire（依赖注入代码生成）

**基础设施**：
- PostgreSQL 16（主数据库）
- Redis 7（缓存和队列）
- MinIO（对象存储）

**可观测性**：
- uber-go/zap（结构化日志）
- OpenTelemetry（分布式追踪）
- Prometheus（指标监控）
- Sentry（错误监控）

## Architecture

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                            │
│         (Flutter App / Admin Panel / Third Party)            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   API Gateway Layer                          │
│    (Load Balancer + Rate Limiter + CORS + Compression)      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  Middleware Pipeline                         │
│  (Auth + Validation + Logging + Error Recovery + Timeout)   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   Controller Layer                           │
│         (Request Handling + Response Formatting)             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
│        (Business Logic + Transaction Management)             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer                            │
│              (Data Access + Query Building)                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                     Model Layer                              │
│              (Domain Models + Relationships)                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                          │
│  (PostgreSQL + Redis + MinIO + Queue + Scheduler + WS)      │
└─────────────────────────────────────────────────────────────┘
```


### 优化后的目录结构

```
.
├── app/
│   ├── http/
│   │   ├── controllers/          # HTTP 控制器
│   │   ├── middleware/           # HTTP 中间件
│   │   ├── requests/             # 请求验证
│   │   └── resources/            # 响应资源转换器
│   ├── services/                 # 业务逻辑层
│   ├── repositories/             # 数据访问层
│   │   ├── interfaces/           # Repository 接口
│   │   └── implementations/      # Repository 实现
│   ├── models/                   # 数据模型
│   ├── dto/                      # 数据传输对象
│   ├── events/                   # 事件定义
│   ├── listeners/                # 事件监听器
│   ├── policies/                 # 授权策略
│   ├── jobs/                     # 队列任务
│   ├── exceptions/               # 自定义异常
│   └── helpers/                  # 辅助函数
├── bootstrap/                    # 应用启动
│   ├── app.go                    # 应用初始化
│   ├── providers.go              # 服务提供者注册
│   └── wire.go                   # Wire 依赖注入配置
├── config/                       # 配置文件
│   ├── app.yaml                  # 应用配置
│   ├── database.yaml             # 数据库配置
│   ├── cache.yaml                # 缓存配置
│   ├── queue.yaml                # 队列配置
│   ├── logging.yaml              # 日志配置
│   └── config.go                 # 配置加载器
├── database/
│   ├── migrations/               # 数据库迁移
│   ├── seeders/                  # 数据填充
│   └── factories/                # 测试数据工厂
├── pkg/                          # 可复用包
│   ├── cache/                    # 缓存包
│   ├── logger/                   # 日志包
│   ├── validator/                # 验证包
│   ├── response/                 # 响应包
│   ├── pagination/               # 分页包
│   ├── jwt/                      # JWT 包
│   ├── hash/                     # 哈希包
│   └── container/                # 服务容器
├── routes/                       # 路由定义
│   ├── api.go                    # API 路由
│   ├── web.go                    # Web 路由
│   └── websocket.go              # WebSocket 路由
├── storage/                      # 存储目录
│   ├── logs/                     # 日志文件
│   ├── cache/                    # 缓存文件
│   └── uploads/                  # 上传文件
├── tests/                        # 测试文件
│   ├── unit/                     # 单元测试
│   ├── integration/              # 集成测试
│   ├── e2e/                      # 端到端测试
│   └── fixtures/                 # 测试数据
├── docs/                         # 文档
│   ├── api/                      # API 文档
│   ├── architecture/             # 架构文档
│   └── deployment/               # 部署文档
├── scripts/                      # 脚本
│   ├── deploy.sh                 # 部署脚本
│   ├── backup.sh                 # 备份脚本
│   └── migrate.sh                # 迁移脚本
├── deployments/                  # 部署配置
│   ├── docker/                   # Docker 配置
│   └── kubernetes/               # K8s 配置
├── cmd/                          # 命令行工具
│   └── server/                   # 服务器入口
│       └── main.go
├── .env.example                  # 环境变量示例
├── Dockerfile                    # Docker 镜像
├── docker-compose.yml            # Docker Compose
├── Makefile                      # Make 命令
├── go.mod                        # Go 模块
└── README.md                     # 项目说明
```

## Components and Interfaces

### 1. 服务容器（Service Container）

服务容器是整个应用的核心，负责管理所有依赖关系。

**接口设计**：
```go
// pkg/container/container.go
type Container interface {
    // 注册服务
    Bind(name string, resolver interface{})
    Singleton(name string, resolver interface{})
    
    // 解析服务
    Make(name string) (interface{}, error)
    MustMake(name string) interface{}
    
    // 检查服务
    Has(name string) bool
    
    // 标签管理
    Tag(tags []string, names []string)
    Tagged(tag string) []interface{}
}

// 实现
type container struct {
    bindings   map[string]*binding
    instances  map[string]interface{}
    tags       map[string][]string
    mu         sync.RWMutex
}

type binding struct {
    resolver  interface{}
    singleton bool
    instance  interface{}
}
```

### 2. 服务提供者（Service Provider）

服务提供者负责注册和启动服务。

**接口设计**：
```go
// bootstrap/provider.go
type ServiceProvider interface {
    // 注册服务到容器
    Register(container Container) error
    
    // 启动服务（在所有服务注册后调用）
    Boot(container Container) error
}

// 示例：数据库服务提供者
type DatabaseServiceProvider struct{}

func (p *DatabaseServiceProvider) Register(c Container) error {
    c.Singleton("db", func() (*gorm.DB, error) {
        cfg := config.Get("database")
        return gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
    })
    return nil
}

func (p *DatabaseServiceProvider) Boot(c Container) error {
    db := c.MustMake("db").(*gorm.DB)
    return db.AutoMigrate(&models.User{}, &models.Order{})
}
```


### 3. 仓储模式（Repository Pattern）

仓储模式封装所有数据访问逻辑，提供统一的数据操作接口。

**接口设计**：
```go
// app/repositories/interfaces/base_repository.go
type BaseRepository[T any] interface {
    // 基础 CRUD
    Find(id uint) (*T, error)
    FindOrFail(id uint) (*T, error)
    All() ([]*T, error)
    Create(entity *T) error
    Update(entity *T) error
    Delete(id uint) error
    
    // 查询构建器
    Where(query interface{}, args ...interface{}) BaseRepository[T]
    WhereIn(column string, values []interface{}) BaseRepository[T]
    OrderBy(column string, direction string) BaseRepository[T]
    Limit(limit int) BaseRepository[T]
    Offset(offset int) BaseRepository[T]
    
    // 分页
    Paginate(page, perPage int) (*pagination.Result[T], error)
    
    // 关联加载
    With(relations ...string) BaseRepository[T]
    
    // 聚合
    Count() (int64, error)
    Exists() (bool, error)
    
    // 事务
    Transaction(fn func(repo BaseRepository[T]) error) error
}

// app/repositories/interfaces/user_repository.go
type UserRepository interface {
    BaseRepository[models.User]
    
    // 自定义方法
    FindByEmail(email string) (*models.User, error)
    FindByEmailOrFail(email string) (*models.User, error)
    UpdatePassword(userID uint, hashedPassword string) error
    GetActiveUsers() ([]*models.User, error)
}

// app/repositories/implementations/user_repository.go
type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err == gorm.ErrRecordNotFound {
        return nil, nil
    }
    return &user, err
}

func (r *userRepository) FindByEmailOrFail(email string) (*models.User, error) {
    user, err := r.FindByEmail(email)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, exceptions.NewNotFoundException("User not found")
    }
    return user, nil
}
```

### 4. 服务层（Service Layer）

服务层封装业务逻辑，协调多个仓储和外部服务。

**接口设计**：
```go
// app/services/auth_service.go
type AuthService interface {
    Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
    Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
    RefreshToken(refreshToken string) (*dto.AuthResponse, error)
    Logout(userID uint) error
    VerifyToken(token string) (*jwt.Claims, error)
}

type authService struct {
    userRepo     repositories.UserRepository
    jwtService   jwt.Service
    hashService  hash.Service
    cacheService cache.Service
    logger       logger.Logger
}

func NewAuthService(
    userRepo repositories.UserRepository,
    jwtService jwt.Service,
    hashService hash.Service,
    cacheService cache.Service,
    logger logger.Logger,
) AuthService {
    return &authService{
        userRepo:     userRepo,
        jwtService:   jwtService,
        hashService:  hashService,
        cacheService: cacheService,
        logger:       logger,
    }
}

func (s *authService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
    // 检查邮箱是否已存在
    existing, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, err
    }
    if existing != nil {
        return nil, exceptions.NewConflictException("Email already exists")
    }
    
    // 哈希密码
    hashedPassword, err := s.hashService.Make(req.Password)
    if err != nil {
        return nil, err
    }
    
    // 创建用户
    user := &models.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashedPassword,
        Role:     "user",
    }
    
    if err := s.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // 生成 Token
    accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Role)
    if err != nil {
        return nil, err
    }
    
    refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
    if err != nil {
        return nil, err
    }
    
    // 记录日志
    s.logger.Info("User registered", 
        logger.Field("user_id", user.ID),
        logger.Field("email", user.Email),
    )
    
    return &dto.AuthResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    3600,
        User:         user,
    }, nil
}
```

### 5. 控制器层（Controller Layer）

控制器负责处理 HTTP 请求，调用服务层，返回响应。

**接口设计**：
```go
// app/http/controllers/auth_controller.go
type AuthController struct {
    authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
    return &AuthController{authService: authService}
}

// Register godoc
// @Summary      用户注册
// @Description  创建新用户账号
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "注册信息"
// @Success      201 {object} response.Success{data=dto.AuthResponse}
// @Failure      400 {object} response.Error
// @Failure      422 {object} response.ValidationError
// @Router       /api/auth/register [post]
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
    // 解析请求
    req := new(dto.RegisterRequest)
    if err := c.BodyParser(req); err != nil {
        return response.BadRequest(c, "Invalid request body")
    }
    
    // 验证请求（由中间件自动处理）
    
    // 调用服务
    result, err := ctrl.authService.Register(req)
    if err != nil {
        return err // 由错误处理中间件处理
    }
    
    // 返回响应
    return response.Created(c, "User registered successfully", result)
}

// Login godoc
// @Summary      用户登录
// @Description  使用邮箱和密码登录
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "登录信息"
// @Success      200 {object} response.Success{data=dto.AuthResponse}
// @Failure      401 {object} response.Error
// @Failure      422 {object} response.ValidationError
// @Router       /api/auth/login [post]
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
    req := new(dto.LoginRequest)
    if err := c.BodyParser(req); err != nil {
        return response.BadRequest(c, "Invalid request body")
    }
    
    result, err := ctrl.authService.Login(req)
    if err != nil {
        return err
    }
    
    return response.Success(c, "Login successful", result)
}
```


### 6. 请求验证（Request Validation）

请求验证使用 validator 库和自定义验证规则。

**接口设计**：
```go
// app/http/requests/register_request.go
type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email,unique:users"`
    Password string `json:"password" validate:"required,min=8,password"`
}

// 自定义验证规则
func (r *RegisterRequest) Rules() map[string]string {
    return map[string]string{
        "name":     "required|min:2|max:50",
        "email":    "required|email|unique:users,email",
        "password": "required|min:8|password",
    }
}

// 自定义错误消息
func (r *RegisterRequest) Messages() map[string]string {
    return map[string]string{
        "name.required":     "姓名不能为空",
        "name.min":          "姓名至少需要 2 个字符",
        "email.required":    "邮箱不能为空",
        "email.email":       "邮箱格式不正确",
        "email.unique":      "邮箱已被注册",
        "password.required": "密码不能为空",
        "password.min":      "密码至少需要 8 个字符",
        "password.password": "密码必须包含大小写字母和数字",
    }
}

// pkg/validator/validator.go
type Validator interface {
    Validate(data interface{}) error
    ValidateStruct(data interface{}) error
    RegisterCustomRule(tag string, fn func(fl validator.FieldLevel) bool)
}

type validator struct {
    validate *validator.Validate
    db       *gorm.DB
}

// 注册自定义验证规则
func (v *validator) RegisterCustomRules() {
    // unique 规则
    v.validate.RegisterValidation("unique", func(fl validator.FieldLevel) bool {
        params := strings.Split(fl.Param(), ",")
        table := params[0]
        column := params[1]
        
        var count int64
        v.db.Table(table).Where(column+" = ?", fl.Field().String()).Count(&count)
        return count == 0
    })
    
    // password 规则（必须包含大小写字母和数字）
    v.validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
        password := fl.Field().String()
        hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
        hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
        hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
        return hasUpper && hasLower && hasNumber
    })
}
```

### 7. 响应格式化（Response Formatting）

统一的 API 响应格式。

**接口设计**：
```go
// pkg/response/response.go
type Response struct {
    Success bool        `json:"success"`
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Errors  interface{} `json:"errors,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    RequestID   string  `json:"request_id"`
    Timestamp   int64   `json:"timestamp"`
    Version     string  `json:"version"`
    Latency     float64 `json:"latency,omitempty"`
}

// 成功响应
func Success(c *fiber.Ctx, message string, data interface{}) error {
    return c.Status(fiber.StatusOK).JSON(Response{
        Success: true,
        Code:    fiber.StatusOK,
        Message: message,
        Data:    data,
        Meta:    buildMeta(c),
    })
}

// 创建成功响应
func Created(c *fiber.Ctx, message string, data interface{}) error {
    return c.Status(fiber.StatusCreated).JSON(Response{
        Success: true,
        Code:    fiber.StatusCreated,
        Message: message,
        Data:    data,
        Meta:    buildMeta(c),
    })
}

// 错误响应
func Error(c *fiber.Ctx, code int, message string, errors interface{}) error {
    return c.Status(code).JSON(Response{
        Success: false,
        Code:    code,
        Message: message,
        Errors:  errors,
        Meta:    buildMeta(c),
    })
}

// 分页响应
type PaginatedResponse struct {
    Items []interface{}   `json:"items"`
    Meta  *PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
    CurrentPage int   `json:"current_page"`
    PerPage     int   `json:"per_page"`
    LastPage    int   `json:"last_page"`
    Total       int64 `json:"total"`
    From        int   `json:"from"`
    To          int   `json:"to"`
    HasMore     bool  `json:"has_more"`
}

func Paginated(c *fiber.Ctx, result *pagination.Result) error {
    return c.JSON(Response{
        Success: true,
        Code:    fiber.StatusOK,
        Message: "Success",
        Data: PaginatedResponse{
            Items: result.Items,
            Meta:  buildPaginationMeta(result),
        },
        Meta: buildMeta(c),
    })
}
```

### 8. 资源转换器（Resource Transformer）

资源转换器用于格式化模型数据为 API 响应。

**接口设计**：
```go
// app/http/resources/user_resource.go
type UserResource struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func NewUserResource(user *models.User) *UserResource {
    return &UserResource{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Role:      user.Role,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
}

func NewUserCollection(users []*models.User) []*UserResource {
    resources := make([]*UserResource, len(users))
    for i, user := range users {
        resources[i] = NewUserResource(user)
    }
    return resources
}

// app/http/resources/order_resource.go
type OrderResource struct {
    ID          uint                `json:"id"`
    OrderNumber string              `json:"order_number"`
    TotalAmount int                 `json:"total_amount"`
    Status      string              `json:"status"`
    QRCode      string              `json:"qr_code"`
    User        *UserResource       `json:"user,omitempty"`
    Device      *DeviceResource     `json:"device,omitempty"`
    Items       []*OrderItemResource `json:"items,omitempty"`
    CreatedAt   time.Time           `json:"created_at"`
    UpdatedAt   time.Time           `json:"updated_at"`
}

func NewOrderResource(order *models.Order, includes ...string) *OrderResource {
    resource := &OrderResource{
        ID:          order.ID,
        OrderNumber: order.OrderNumber,
        TotalAmount: order.TotalAmount,
        Status:      order.Status,
        QRCode:      order.QRCode,
        CreatedAt:   order.CreatedAt,
        UpdatedAt:   order.UpdatedAt,
    }
    
    // 条件加载关联数据
    for _, include := range includes {
        switch include {
        case "user":
            if order.User.ID != 0 {
                resource.User = NewUserResource(&order.User)
            }
        case "device":
            if order.Device.ID != 0 {
                resource.Device = NewDeviceResource(&order.Device)
            }
        case "items":
            resource.Items = NewOrderItemCollection(order.Items)
        }
    }
    
    return resource
}
```


## Data Models

### 数据模型设计

所有模型继承基础模型，提供统一的字段和方法。

```go
// app/models/base_model.go
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// 模型事件钩子
type ModelEvents interface {
    BeforeCreate(tx *gorm.DB) error
    AfterCreate(tx *gorm.DB) error
    BeforeUpdate(tx *gorm.DB) error
    AfterUpdate(tx *gorm.DB) error
    BeforeDelete(tx *gorm.DB) error
    AfterDelete(tx *gorm.DB) error
}

// app/models/user.go
type User struct {
    BaseModel
    Name     string  `gorm:"type:varchar(100);not null" json:"name"`
    Email    string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
    Password string  `gorm:"type:varchar(255);not null" json:"-"`
    Role     string  `gorm:"type:varchar(20);default:'user'" json:"role"`
    Status   string  `gorm:"type:varchar(20);default:'active'" json:"status"`
    Orders   []Order `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}

// 表名
func (User) TableName() string {
    return "users"
}

// 模型事件
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 验证邮箱格式
    if !isValidEmail(u.Email) {
        return errors.New("invalid email format")
    }
    return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
    // 触发用户创建事件
    events.Dispatch("user.created", u)
    return nil
}

// app/models/order.go
type Order struct {
    BaseModel
    UserID      uint       `gorm:"not null;index" json:"user_id"`
    User        User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
    DeviceID    uint       `gorm:"not null;index" json:"device_id"`
    Device      Device     `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
    OrderNumber string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_number"`
    TotalAmount int        `gorm:"not null" json:"total_amount"`
    Status      string     `gorm:"type:varchar(20);default:'pending';index" json:"status"`
    QRCode      string     `gorm:"type:varchar(255);uniqueIndex" json:"qr_code"`
    PaidAt      *time.Time `json:"paid_at,omitempty"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
    Payment     *Payment   `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
}

func (Order) TableName() string {
    return "orders"
}

// 作用域
func (o *Order) ScopePending(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "pending")
}

func (o *Order) ScopePaid(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "paid")
}

func (o *Order) ScopeCompleted(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "completed")
}

// 业务方法
func (o *Order) IsPending() bool {
    return o.Status == "pending"
}

func (o *Order) IsPaid() bool {
    return o.Status == "paid"
}

func (o *Order) CanBeCancelled() bool {
    return o.Status == "pending" || o.Status == "paid"
}

func (o *Order) MarkAsPaid() {
    now := time.Now()
    o.Status = "paid"
    o.PaidAt = &now
}

func (o *Order) MarkAsCompleted() {
    now := time.Now()
    o.Status = "completed"
    o.CompletedAt = &now
}
```

### 数据库迁移

使用 GORM AutoMigrate 和自定义迁移文件。

```go
// database/migrations/000001_create_users_table.go
type CreateUsersTable struct{}

func (m *CreateUsersTable) Up(db *gorm.DB) error {
    return db.AutoMigrate(&models.User{})
}

func (m *CreateUsersTable) Down(db *gorm.DB) error {
    return db.Migrator().DropTable(&models.User{})
}

// database/migrations/000002_add_indexes.go
type AddIndexes struct{}

func (m *AddIndexes) Up(db *gorm.DB) error {
    // 添加复合索引
    if err := db.Exec(`
        CREATE INDEX idx_orders_user_status ON orders(user_id, status)
    `).Error; err != nil {
        return err
    }
    
    if err := db.Exec(`
        CREATE INDEX idx_orders_device_created ON orders(device_id, created_at DESC)
    `).Error; err != nil {
        return err
    }
    
    return nil
}

func (m *AddIndexes) Down(db *gorm.DB) error {
    db.Exec(`DROP INDEX IF EXISTS idx_orders_user_status`)
    db.Exec(`DROP INDEX IF EXISTS idx_orders_device_created`)
    return nil
}
```

## Error Handling

### 错误类型层次结构

```go
// app/exceptions/base_exception.go
type BaseException struct {
    Message    string
    Code       int
    StatusCode int
    Errors     map[string][]string
    Cause      error
}

func (e *BaseException) Error() string {
    return e.Message
}

func (e *BaseException) Unwrap() error {
    return e.Cause
}

// app/exceptions/http_exceptions.go
type ValidationException struct {
    *BaseException
}

func NewValidationException(message string, errors map[string][]string) *ValidationException {
    return &ValidationException{
        BaseException: &BaseException{
            Message:    message,
            Code:       "VALIDATION_ERROR",
            StatusCode: fiber.StatusUnprocessableEntity,
            Errors:     errors,
        },
    }
}

type AuthenticationException struct {
    *BaseException
}

func NewAuthenticationException(message string) *AuthenticationException {
    return &AuthenticationException{
        BaseException: &BaseException{
            Message:    message,
            Code:       "AUTHENTICATION_ERROR",
            StatusCode: fiber.StatusUnauthorized,
        },
    }
}

type NotFoundException struct {
    *BaseException
}

func NewNotFoundException(message string) *NotFoundException {
    return &NotFoundException{
        BaseException: &BaseException{
            Message:    message,
            Code:       "NOT_FOUND",
            StatusCode: fiber.StatusNotFound,
        },
    }
}

type ConflictException struct {
    *BaseException
}

func NewConflictException(message string) *ConflictException {
    return &ConflictException{
        BaseException: &BaseException{
            Message:    message,
            Code:       "CONFLICT",
            StatusCode: fiber.StatusConflict,
        },
    }
}
```

### 全局错误处理中间件

```go
// app/http/middleware/error_handler.go
func ErrorHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        err := c.Next()
        if err == nil {
            return nil
        }
        
        // 获取 logger 和 request ID
        logger := c.Locals("logger").(logger.Logger)
        requestID := c.Locals("request_id").(string)
        
        // 处理自定义异常
        var baseErr *exceptions.BaseException
        if errors.As(err, &baseErr) {
            logger.Error("Application error",
                logger.Field("request_id", requestID),
                logger.Field("error", baseErr.Message),
                logger.Field("code", baseErr.Code),
            )
            
            response := fiber.Map{
                "success": false,
                "code":    baseErr.StatusCode,
                "message": baseErr.Message,
                "errors":  baseErr.Errors,
                "meta": fiber.Map{
                    "request_id": requestID,
                    "timestamp":  time.Now().Unix(),
                },
            }
            
            // 开发环境添加调试信息
            if config.Get("app.debug").(bool) {
                response["debug"] = fiber.Map{
                    "exception": fmt.Sprintf("%T", err),
                    "trace":     getStackTrace(err),
                }
            }
            
            return c.Status(baseErr.StatusCode).JSON(response)
        }
        
        // 处理 Fiber 错误
        var fiberErr *fiber.Error
        if errors.As(err, &fiberErr) {
            return c.Status(fiberErr.Code).JSON(fiber.Map{
                "success": false,
                "code":    fiberErr.Code,
                "message": fiberErr.Message,
                "meta": fiber.Map{
                    "request_id": requestID,
                    "timestamp":  time.Now().Unix(),
                },
            })
        }
        
        // 处理未知错误
        logger.Error("Unexpected error",
            logger.Field("request_id", requestID),
            logger.Field("error", err.Error()),
            logger.Field("stack", getStackTrace(err)),
        )
        
        // 发送到 Sentry
        sentry.CaptureException(err)
        
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "code":    fiber.StatusInternalServerError,
            "message": "Internal server error",
            "meta": fiber.Map{
                "request_id": requestID,
                "timestamp":  time.Now().Unix(),
            },
        })
    }
}
```


## Testing Strategy

### 测试架构

采用多层次测试策略：单元测试、集成测试、端到端测试。

**测试工具**：
- `testing`：Go 标准测试库
- `testify`：断言和 Mock 库
- `gomock`：Mock 生成工具
- `httptest`：HTTP 测试工具
- `dockertest`：Docker 容器测试

### 单元测试

测试单个函数和方法，使用 Mock 隔离依赖。

```go
// app/services/auth_service_test.go
type AuthServiceTestSuite struct {
    suite.Suite
    mockUserRepo     *mocks.MockUserRepository
    mockJWTService   *mocks.MockJWTService
    mockHashService  *mocks.MockHashService
    mockCacheService *mocks.MockCacheService
    mockLogger       *mocks.MockLogger
    authService      services.AuthService
}

func (suite *AuthServiceTestSuite) SetupTest() {
    ctrl := gomock.NewController(suite.T())
    suite.mockUserRepo = mocks.NewMockUserRepository(ctrl)
    suite.mockJWTService = mocks.NewMockJWTService(ctrl)
    suite.mockHashService = mocks.NewMockHashService(ctrl)
    suite.mockCacheService = mocks.NewMockCacheService(ctrl)
    suite.mockLogger = mocks.NewMockLogger(ctrl)
    
    suite.authService = services.NewAuthService(
        suite.mockUserRepo,
        suite.mockJWTService,
        suite.mockHashService,
        suite.mockCacheService,
        suite.mockLogger,
    )
}

func (suite *AuthServiceTestSuite) TestRegister_Success() {
    // Arrange
    req := &dto.RegisterRequest{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "Password123",
    }
    
    suite.mockUserRepo.EXPECT().
        FindByEmail(req.Email).
        Return(nil, nil)
    
    suite.mockHashService.EXPECT().
        Make(req.Password).
        Return("hashed_password", nil)
    
    suite.mockUserRepo.EXPECT().
        Create(gomock.Any()).
        Return(nil)
    
    suite.mockJWTService.EXPECT().
        GenerateAccessToken(gomock.Any(), "user").
        Return("access_token", nil)
    
    suite.mockJWTService.EXPECT().
        GenerateRefreshToken(gomock.Any()).
        Return("refresh_token", nil)
    
    suite.mockLogger.EXPECT().
        Info(gomock.Any(), gomock.Any()).
        Times(1)
    
    // Act
    result, err := suite.authService.Register(req)
    
    // Assert
    suite.NoError(err)
    suite.NotNil(result)
    suite.Equal("access_token", result.AccessToken)
    suite.Equal("refresh_token", result.RefreshToken)
}

func (suite *AuthServiceTestSuite) TestRegister_EmailExists() {
    // Arrange
    req := &dto.RegisterRequest{
        Name:     "Test User",
        Email:    "existing@example.com",
        Password: "Password123",
    }
    
    existingUser := &models.User{
        BaseModel: models.BaseModel{ID: 1},
        Email:     req.Email,
    }
    
    suite.mockUserRepo.EXPECT().
        FindByEmail(req.Email).
        Return(existingUser, nil)
    
    // Act
    result, err := suite.authService.Register(req)
    
    // Assert
    suite.Error(err)
    suite.Nil(result)
    suite.IsType(&exceptions.ConflictException{}, err)
}

func TestAuthServiceTestSuite(t *testing.T) {
    suite.Run(t, new(AuthServiceTestSuite))
}
```

### 集成测试

测试多个组件的集成，使用真实数据库。

```go
// tests/integration/auth_test.go
type AuthIntegrationTestSuite struct {
    suite.Suite
    app         *fiber.App
    db          *gorm.DB
    container   *dockertest.Pool
    resource    *dockertest.Resource
}

func (suite *AuthIntegrationTestSuite) SetupSuite() {
    // 启动测试数据库容器
    pool, err := dockertest.NewPool("")
    suite.Require().NoError(err)
    
    resource, err := pool.Run("postgres", "16", []string{
        "POSTGRES_PASSWORD=test",
        "POSTGRES_DB=test",
    })
    suite.Require().NoError(err)
    
    suite.container = pool
    suite.resource = resource
    
    // 连接数据库
    var db *gorm.DB
    err = pool.Retry(func() error {
        dsn := fmt.Sprintf(
            "host=localhost port=%s user=postgres password=test dbname=test sslmode=disable",
            resource.GetPort("5432/tcp"),
        )
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        return err
    })
    suite.Require().NoError(err)
    suite.db = db
    
    // 运行迁移
    err = db.AutoMigrate(&models.User{}, &models.Order{})
    suite.Require().NoError(err)
    
    // 初始化应用
    suite.app = bootstrap.NewApp(db)
}

func (suite *AuthIntegrationTestSuite) TearDownSuite() {
    suite.container.Purge(suite.resource)
}

func (suite *AuthIntegrationTestSuite) SetupTest() {
    // 清理数据
    suite.db.Exec("TRUNCATE users CASCADE")
}

func (suite *AuthIntegrationTestSuite) TestRegisterAndLogin() {
    // 注册用户
    registerReq := map[string]string{
        "name":     "Test User",
        "email":    "test@example.com",
        "password": "Password123",
    }
    
    registerBody, _ := json.Marshal(registerReq)
    req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(registerBody))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := suite.app.Test(req)
    suite.NoError(err)
    suite.Equal(fiber.StatusCreated, resp.StatusCode)
    
    var registerResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&registerResp)
    suite.True(registerResp["success"].(bool))
    
    // 登录用户
    loginReq := map[string]string{
        "email":    "test@example.com",
        "password": "Password123",
    }
    
    loginBody, _ := json.Marshal(loginReq)
    req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err = suite.app.Test(req)
    suite.NoError(err)
    suite.Equal(fiber.StatusOK, resp.StatusCode)
    
    var loginResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&loginResp)
    suite.True(loginResp["success"].(bool))
    
    data := loginResp["data"].(map[string]interface{})
    suite.NotEmpty(data["access_token"])
    suite.NotEmpty(data["refresh_token"])
}

func TestAuthIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(AuthIntegrationTestSuite))
}
```

### 性能测试

使用基准测试评估性能。

```go
// app/services/auth_service_bench_test.go
func BenchmarkAuthService_Register(b *testing.B) {
    // Setup
    db := setupTestDB()
    authService := setupAuthService(db)
    
    req := &dto.RegisterRequest{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "Password123",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        req.Email = fmt.Sprintf("test%d@example.com", i)
        authService.Register(req)
    }
}

func BenchmarkAuthService_Login(b *testing.B) {
    // Setup
    db := setupTestDB()
    authService := setupAuthService(db)
    
    // 创建测试用户
    user := createTestUser(db)
    
    req := &dto.LoginRequest{
        Email:    user.Email,
        Password: "Password123",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        authService.Login(req)
    }
}
```

## 核心功能设计

### 1. 配置管理

使用 Viper 实现灵活的配置管理。

```go
// pkg/config/config.go
type Config struct {
    v *viper.Viper
}

func New() *Config {
    v := viper.New()
    
    // 设置配置文件
    v.SetConfigName("app")
    v.SetConfigType("yaml")
    v.AddConfigPath("./config")
    v.AddConfigPath(".")
    
    // 环境变量
    v.SetEnvPrefix("APP")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // 读取配置
    if err := v.ReadInConfig(); err != nil {
        panic(err)
    }
    
    // 监听配置变化
    v.WatchConfig()
    v.OnConfigChange(func(e fsnotify.Event) {
        log.Println("Config file changed:", e.Name)
    })
    
    return &Config{v: v}
}

func (c *Config) Get(key string) interface{} {
    return c.v.Get(key)
}

func (c *Config) GetString(key string) string {
    return c.v.GetString(key)
}

func (c *Config) GetInt(key string) int {
    return c.v.GetInt(key)
}

func (c *Config) GetBool(key string) bool {
    return c.v.GetBool(key)
}

// 配置验证
func (c *Config) Validate() error {
    required := []string{
        "app.name",
        "app.port",
        "database.host",
        "database.port",
        "database.name",
        "jwt.secret",
    }
    
    for _, key := range required {
        if !c.v.IsSet(key) {
            return fmt.Errorf("required config key missing: %s", key)
        }
    }
    
    return nil
}
```


### 2. 缓存系统

实现多级缓存和缓存策略。

```go
// pkg/cache/cache.go
type Cache interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    DeletePattern(ctx context.Context, pattern string) error
    Has(ctx context.Context, key string) (bool, error)
    Remember(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error)
    Tags(tags ...string) Cache
    Flush(ctx context.Context) error
}

// 二级缓存实现
type tieredCache struct {
    l1 Cache // Memory cache
    l2 Cache // Redis cache
}

func NewTieredCache(l1, l2 Cache) Cache {
    return &tieredCache{l1: l1, l2: l2}
}

func (c *tieredCache) Get(ctx context.Context, key string, dest interface{}) error {
    // 先从 L1 获取
    err := c.l1.Get(ctx, key, dest)
    if err == nil {
        return nil
    }
    
    // L1 未命中，从 L2 获取
    err = c.l2.Get(ctx, key, dest)
    if err != nil {
        return err
    }
    
    // 回填 L1
    c.l1.Set(ctx, key, dest, 5*time.Minute)
    return nil
}

func (c *tieredCache) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
    // 尝试从缓存获取
    var result interface{}
    err := c.Get(ctx, key, &result)
    if err == nil {
        return result, nil
    }
    
    // 使用互斥锁防止缓存击穿
    mu := sync.Mutex{}
    mu.Lock()
    defer mu.Unlock()
    
    // 双重检查
    err = c.Get(ctx, key, &result)
    if err == nil {
        return result, nil
    }
    
    // 执行回调函数
    result, err = fn()
    if err != nil {
        return nil, err
    }
    
    // 存入缓存
    c.Set(ctx, key, result, ttl)
    return result, nil
}

// Redis 缓存实现
type redisCache struct {
    client *redis.Client
    prefix string
    tags   []string
}

func NewRedisCache(client *redis.Client, prefix string) Cache {
    return &redisCache{
        client: client,
        prefix: prefix,
    }
}

func (c *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, c.buildKey(key)).Result()
    if err == redis.Nil {
        return ErrCacheNotFound
    }
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(val), dest)
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    // 添加随机过期时间防止缓存雪崩
    jitter := time.Duration(rand.Intn(60)) * time.Second
    ttl = ttl + jitter
    
    err = c.client.Set(ctx, c.buildKey(key), data, ttl).Err()
    if err != nil {
        return err
    }
    
    // 如果有标签，添加到标签集合
    if len(c.tags) > 0 {
        for _, tag := range c.tags {
            c.client.SAdd(ctx, c.tagKey(tag), c.buildKey(key))
        }
    }
    
    return nil
}

func (c *redisCache) Tags(tags ...string) Cache {
    return &redisCache{
        client: c.client,
        prefix: c.prefix,
        tags:   tags,
    }
}

func (c *redisCache) DeletePattern(ctx context.Context, pattern string) error {
    iter := c.client.Scan(ctx, 0, c.buildKey(pattern), 0).Iterator()
    for iter.Next(ctx) {
        c.client.Del(ctx, iter.Val())
    }
    return iter.Err()
}
```

### 3. 日志系统

使用 Zap 实现结构化日志。

```go
// pkg/logger/logger.go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    With(fields ...Field) Logger
}

type Field struct {
    Key   string
    Value interface{}
}

func NewField(key string, value interface{}) Field {
    return Field{Key: key, Value: value}
}

type zapLogger struct {
    logger *zap.Logger
}

func NewZapLogger(config *Config) Logger {
    var zapConfig zap.Config
    
    if config.GetBool("app.debug") {
        zapConfig = zap.NewDevelopmentConfig()
    } else {
        zapConfig = zap.NewProductionConfig()
    }
    
    // 配置输出
    zapConfig.OutputPaths = []string{
        "stdout",
        config.GetString("logging.file"),
    }
    
    // 配置日志轮转
    zapConfig.OutputPaths = append(zapConfig.OutputPaths, 
        fmt.Sprintf("lumberjack:%s", config.GetString("logging.file")))
    
    logger, err := zapConfig.Build(
        zap.AddCaller(),
        zap.AddStacktrace(zap.ErrorLevel),
    )
    if err != nil {
        panic(err)
    }
    
    return &zapLogger{logger: logger}
}

func (l *zapLogger) Info(msg string, fields ...Field) {
    l.logger.Info(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
    l.logger.Error(msg, l.convertFields(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
    return &zapLogger{
        logger: l.logger.With(l.convertFields(fields)...),
    }
}

func (l *zapLogger) convertFields(fields []Field) []zap.Field {
    zapFields := make([]zap.Field, len(fields))
    for i, f := range fields {
        zapFields[i] = zap.Any(f.Key, f.Value)
    }
    return zapFields
}

// 日志中间件
func LoggerMiddleware(logger Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        requestID := uuid.New().String()
        
        // 设置 request ID
        c.Locals("request_id", requestID)
        
        // 创建请求日志器
        reqLogger := logger.With(
            Field{"request_id", requestID},
            Field{"method", c.Method()},
            Field{"path", c.Path()},
            Field{"ip", c.IP()},
        )
        c.Locals("logger", reqLogger)
        
        // 处理请求
        err := c.Next()
        
        // 记录响应日志
        latency := time.Since(start)
        reqLogger.Info("Request completed",
            Field{"status", c.Response().StatusCode()},
            Field{"latency", latency.Milliseconds()},
            Field{"size", len(c.Response().Body())},
        )
        
        return err
    }
}
```

### 4. 队列系统

使用 Asynq 实现可靠的队列系统。

```go
// pkg/queue/queue.go
type Queue interface {
    Dispatch(task Task, opts ...Option) error
    DispatchAfter(task Task, delay time.Duration, opts ...Option) error
    RegisterHandler(taskType string, handler Handler)
    Start() error
    Stop() error
}

type Task interface {
    Type() string
    Payload() ([]byte, error)
}

type Handler func(ctx context.Context, task *asynq.Task) error

type Option func(*asynq.TaskInfo)

// Asynq 队列实现
type asynqQueue struct {
    client *asynq.Client
    server *asynq.Server
    mux    *asynq.ServeMux
}

func NewAsynqQueue(redisOpt asynq.RedisClientOpt) Queue {
    client := asynq.NewClient(redisOpt)
    
    server := asynq.NewServer(redisOpt, asynq.Config{
        Concurrency: 10,
        Queues: map[string]int{
            "critical": 6,
            "default":  3,
            "low":      1,
        },
        RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
            // 指数退避
            return time.Duration(math.Pow(2, float64(n))) * time.Second
        },
    })
    
    return &asynqQueue{
        client: client,
        server: server,
        mux:    asynq.NewServeMux(),
    }
}

func (q *asynqQueue) Dispatch(task Task, opts ...Option) error {
    payload, err := task.Payload()
    if err != nil {
        return err
    }
    
    asynqTask := asynq.NewTask(task.Type(), payload)
    
    // 应用选项
    taskOpts := []asynq.Option{}
    for _, opt := range opts {
        // 转换选项
    }
    
    _, err = q.client.Enqueue(asynqTask, taskOpts...)
    return err
}

func (q *asynqQueue) RegisterHandler(taskType string, handler Handler) {
    q.mux.HandleFunc(taskType, handler)
}

func (q *asynqQueue) Start() error {
    return q.server.Run(q.mux)
}

// 示例任务
type SendEmailTask struct {
    To      string
    Subject string
    Body    string
}

func (t *SendEmailTask) Type() string {
    return "email:send"
}

func (t *SendEmailTask) Payload() ([]byte, error) {
    return json.Marshal(t)
}

// 任务处理器
func HandleSendEmail(emailService services.EmailService) Handler {
    return func(ctx context.Context, task *asynq.Task) error {
        var payload SendEmailTask
        if err := json.Unmarshal(task.Payload(), &payload); err != nil {
            return err
        }
        
        return emailService.Send(payload.To, payload.Subject, payload.Body)
    }
}
```


### 5. 监控和可观测性

集成 Prometheus、OpenTelemetry 和 Sentry。

```go
// pkg/monitoring/metrics.go
type Metrics interface {
    RecordHTTPRequest(method, path string, status int, duration time.Duration)
    RecordDBQuery(query string, duration time.Duration, err error)
    RecordCacheHit(key string)
    RecordCacheMiss(key string)
    RecordQueueTask(taskType string, status string, duration time.Duration)
    IncrementCounter(name string, labels map[string]string)
    RecordHistogram(name string, value float64, labels map[string]string)
    RecordGauge(name string, value float64, labels map[string]string)
}

type prometheusMetrics struct {
    httpRequestsTotal   *prometheus.CounterVec
    httpRequestDuration *prometheus.HistogramVec
    dbQueryDuration     *prometheus.HistogramVec
    dbQueryErrors       *prometheus.CounterVec
    cacheHits           *prometheus.CounterVec
    cacheMisses         *prometheus.CounterVec
    queueTaskDuration   *prometheus.HistogramVec
}

func NewPrometheusMetrics() Metrics {
    m := &prometheusMetrics{
        httpRequestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "path", "status"},
        ),
        httpRequestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "http_request_duration_seconds",
                Help:    "HTTP request duration in seconds",
                Buckets: prometheus.DefBuckets,
            },
            []string{"method", "path"},
        ),
        dbQueryDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "db_query_duration_seconds",
                Help:    "Database query duration in seconds",
                Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
            },
            []string{"query"},
        ),
        cacheHits: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "cache_hits_total",
                Help: "Total number of cache hits",
            },
            []string{"key"},
        ),
    }
    
    // 注册指标
    prometheus.MustRegister(
        m.httpRequestsTotal,
        m.httpRequestDuration,
        m.dbQueryDuration,
        m.cacheHits,
        m.cacheMisses,
    )
    
    return m
}

func (m *prometheusMetrics) RecordHTTPRequest(method, path string, status int, duration time.Duration) {
    m.httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
    m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// Prometheus 中间件
func PrometheusMiddleware(metrics Metrics) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        
        err := c.Next()
        
        duration := time.Since(start)
        metrics.RecordHTTPRequest(
            c.Method(),
            c.Route().Path,
            c.Response().StatusCode(),
            duration,
        )
        
        return err
    }
}

// OpenTelemetry 追踪
func TracingMiddleware(tracer trace.Tracer) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx, span := tracer.Start(c.Context(), c.Route().Path,
            trace.WithAttributes(
                attribute.String("http.method", c.Method()),
                attribute.String("http.url", c.OriginalURL()),
                attribute.String("http.user_agent", c.Get("User-Agent")),
            ),
        )
        defer span.End()
        
        c.SetUserContext(ctx)
        
        err := c.Next()
        
        span.SetAttributes(
            attribute.Int("http.status_code", c.Response().StatusCode()),
        )
        
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
        }
        
        return err
    }
}

// Sentry 错误监控
func SentryMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        hub := sentry.CurrentHub().Clone()
        hub.Scope().SetRequest(c.Request())
        hub.Scope().SetTag("request_id", c.Locals("request_id").(string))
        
        c.Locals("sentry_hub", hub)
        
        err := c.Next()
        
        if err != nil {
            hub.CaptureException(err)
        }
        
        return err
    }
}

// 健康检查
type HealthCheck struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewHealthCheck(db *gorm.DB, redis *redis.Client) *HealthCheck {
    return &HealthCheck{db: db, redis: redis}
}

func (h *HealthCheck) Check(c *fiber.Ctx) error {
    checks := map[string]string{
        "database": h.checkDatabase(),
        "redis":    h.checkRedis(),
    }
    
    allHealthy := true
    for _, status := range checks {
        if status != "healthy" {
            allHealthy = false
            break
        }
    }
    
    status := fiber.StatusOK
    if !allHealthy {
        status = fiber.StatusServiceUnavailable
    }
    
    return c.Status(status).JSON(fiber.Map{
        "status": map[string]bool{"healthy": allHealthy}[strconv.FormatBool(allHealthy)],
        "checks": checks,
        "timestamp": time.Now().Unix(),
    })
}

func (h *HealthCheck) checkDatabase() string {
    sqlDB, err := h.db.DB()
    if err != nil {
        return "unhealthy"
    }
    
    if err := sqlDB.Ping(); err != nil {
        return "unhealthy"
    }
    
    return "healthy"
}

func (h *HealthCheck) checkRedis() string {
    if err := h.redis.Ping(context.Background()).Err(); err != nil {
        return "unhealthy"
    }
    return "healthy"
}
```

### 6. 安全性

实现全面的安全防护。

```go
// app/http/middleware/security.go

// CORS 中间件
func CORSMiddleware() fiber.Handler {
    return cors.New(cors.Config{
        AllowOrigins:     config.GetString("app.cors.origins"),
        AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
        MaxAge:           86400,
    })
}

// 安全头中间件
func SecurityHeadersMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        c.Set("X-Frame-Options", "DENY")
        c.Set("X-Content-Type-Options", "nosniff")
        c.Set("X-XSS-Protection", "1; mode=block")
        c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Set("Content-Security-Policy", "default-src 'self'")
        c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        return c.Next()
    }
}

// 速率限制中间件
func RateLimitMiddleware(limiter *limiter.Limiter) fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := c.IP()
        
        // 对于认证用户，使用用户 ID
        if userID := c.Locals("user_id"); userID != nil {
            key = fmt.Sprintf("user:%v", userID)
        }
        
        allowed, err := limiter.Allow(c.Context(), key)
        if err != nil {
            return err
        }
        
        if !allowed {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "success": false,
                "message": "Too many requests",
            })
        }
        
        return c.Next()
    }
}

// 输入清理中间件
func SanitizeInputMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 清理查询参数
        for key, values := range c.Queries() {
            for i, value := range values {
                values[i] = html.EscapeString(value)
            }
        }
        
        return c.Next()
    }
}

// CSRF 防护
type CSRFProtection struct {
    secret []byte
}

func NewCSRFProtection(secret string) *CSRFProtection {
    return &CSRFProtection{secret: []byte(secret)}
}

func (p *CSRFProtection) Middleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // GET 请求不需要 CSRF 验证
        if c.Method() == "GET" {
            return c.Next()
        }
        
        token := c.Get("X-CSRF-Token")
        if token == "" {
            token = c.FormValue("_csrf")
        }
        
        if !p.validateToken(token) {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "success": false,
                "message": "CSRF token validation failed",
            })
        }
        
        return c.Next()
    }
}

func (p *CSRFProtection) GenerateToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

func (p *CSRFProtection) validateToken(token string) bool {
    // 实现 token 验证逻辑
    return true
}
```

### 7. 部署配置

提供完整的部署配置。

```dockerfile
# deployments/docker/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git make

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

# 暴露端口
EXPOSE 3000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# 运行应用
CMD ["./main"]
```

```yaml
# deployments/kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lunchbox-api
  labels:
    app: lunchbox-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: lunchbox-api
  template:
    metadata:
      labels:
        app: lunchbox-api
    spec:
      containers:
      - name: api
        image: lunchbox-api:latest
        ports:
        - containerPort: 3000
        env:
        - name: APP_ENV
          value: "production"
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: host
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: lunchbox-api
spec:
  selector:
    app: lunchbox-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: LoadBalancer
```

## 性能优化策略

### 1. 数据库优化

- 使用连接池管理数据库连接
- 实现读写分离
- 添加适当的索引
- 使用预加载避免 N+1 查询
- 实现查询结果缓存

### 2. 缓存优化

- 实现二级缓存（Memory + Redis）
- 使用缓存标签批量失效
- 实现缓存预热
- 防止缓存穿透、击穿、雪崩

### 3. API 优化

- 实现响应压缩
- 使用 HTTP/2
- 实现 API 响应缓存
- 使用 CDN 加速静态资源

### 4. 并发优化

- 使用 goroutine 池
- 实现请求合并
- 使用异步处理耗时操作

## 总结

本设计文档提供了一个完整的、生产级别的 Go 框架优化方案，完全参照 Laravel 12 的最佳实践，使用最新的 Go 1.24 和生态系统包。通过实现清晰的分层架构、完善的依赖注入、统一的错误处理、全面的测试策略和强大的监控系统，确保系统具备高性能、高可用和高安全性。
