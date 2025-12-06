# Requirements Document

## Introduction

本文档定义了饭盒售货机后端 RESTful API 系统的需求。该系统为 Flutter 移动应用提供完整的后端支持，实现城市-设备-产品的三级管理架构，支持用户认证、订单管理、支付集成、设备位置上报等核心功能。系统采用 Golang Fiber 框架开发，完全仿照 Laravel 框架的设计理念。

## Glossary

- **System**: 饭盒售货机后端 API 系统
- **User**: 使用移动应用购买产品的终端用户
- **Admin**: 系统管理员，负责管理城市、设备、产品等
- **City**: 城市实体，包含多个设备的地理分区单位
- **Device**: 售货机设备实体，位于特定城市，包含多个产品
- **Product**: 产品实体，位于特定设备中的可购买商品
- **Order**: 订单实体，用户的购买记录
- **Payment**: 支付实体，订单的支付信息
- **JWT Token**: JSON Web Token，用于用户认证的令牌
- **Location**: 地理位置信息，包含经纬度坐标
- **Category**: 产品分类
- **Inventory**: 库存信息
- **Transaction**: 交易记录

## Requirements

### Requirement 1: 用户认证与授权

**User Story:** 作为用户，我希望能够注册账号并登录系统，以便安全地使用售货机服务

#### Acceptance Criteria

1. WHEN 用户提交有效的注册信息（邮箱、密码、姓名）THEN THE System SHALL 创建新用户账号并返回成功响应
2. WHEN 用户提交已存在的邮箱注册 THEN THE System SHALL 拒绝注册并返回错误信息
3. WHEN 用户使用正确的邮箱和密码登录 THEN THE System SHALL 生成 JWT Token 并返回给用户
4. WHEN 用户使用错误的凭证登录 THEN THE System SHALL 拒绝登录并返回认证失败错误
5. WHEN 用户携带有效的 JWT Token 访问受保护的 API THEN THE System SHALL 验证 Token 并允许访问
6. WHEN 用户携带过期的 JWT Token 访问 API THEN THE System SHALL 拒绝访问并返回 Token 过期错误
7. WHEN 用户请求刷新 Token THEN THE System SHALL 验证旧 Token 并生成新的 JWT Token

### Requirement 2: 城市管理

**User Story:** 作为管理员，我希望能够管理城市信息，以便组织和分配售货机设备

#### Acceptance Criteria

1. WHEN 任何用户请求城市列表 THEN THE System SHALL 返回所有城市的基本信息（ID、名称、状态）
2. WHEN 用户请求特定城市详情 THEN THE System SHALL 返回该城市的完整信息和关联的设备数量
3. WHEN 管理员提交有效的城市信息（名称、描述、状态）THEN THE System SHALL 创建新城市记录
4. WHEN 管理员更新城市信息 THEN THE System SHALL 验证权限并更新城市记录
5. WHEN 管理员删除城市 THEN THE System SHALL 检查该城市是否有关联设备，如有则拒绝删除
6. WHEN 非管理员用户尝试创建、更新或删除城市 THEN THE System SHALL 拒绝操作并返回权限不足错误

### Requirement 3: 设备管理

**User Story:** 作为管理员，我希望能够管理售货机设备，以便监控设备状态和库存

#### Acceptance Criteria

1. WHEN 用户请求设备列表 THEN THE System SHALL 返回所有设备的基本信息（ID、名称、位置、状态）
2. WHEN 用户按城市 ID 筛选设备 THEN THE System SHALL 返回该城市下的所有设备
3. WHEN 用户请求特定设备详情 THEN THE System SHALL 返回设备的完整信息（包括位置、状态、库存统计）
4. WHEN 管理员创建新设备 THEN THE System SHALL 验证城市 ID 存在并创建设备记录
5. WHEN 设备上报位置信息（经纬度）THEN THE System SHALL 更新设备的地理位置记录
6. WHEN 设备上报状态信息（在线、离线、故障）THEN THE System SHALL 更新设备状态并记录时间戳
7. WHEN 管理员更新设备信息 THEN THE System SHALL 验证权限并更新设备记录
8. WHEN 管理员删除设备 THEN THE System SHALL 检查该设备是否有未完成订单，如有则拒绝删除

### Requirement 4: 产品管理与货道系统

**User Story:** 作为管理员，我希望能够管理产品信息、设备货道和双库存系统，以便精确控制产品销售和库存

#### Acceptance Criteria

1. THE System SHALL 为每台设备提供 53 个货道（Channel）
2. WHEN 创建设备 THEN THE System SHALL 自动初始化 53 个货道记录（编号 1-53）
3. WHEN 管理员分配产品到货道 THEN THE System SHALL 记录货道编号、产品 ID、虚拟库存、实际库存
4. THE System SHALL 为每个货道维护两种库存：虚拟库存（virtual_stock）和实际库存（actual_stock）
5. WHEN 设置货道库存 THEN THE System SHALL 验证虚拟库存和实际库存的最大值为 4
6. WHEN 用户浏览设备产品 THEN THE System SHALL 返回该设备所有货道的产品信息和虚拟库存
7. WHEN 用户创建订单 THEN THE System SHALL 扣减对应货道的虚拟库存
8. WHEN 用户扫码取货成功 THEN THE System SHALL 扣减对应货道的实际库存
9. WHEN 订单取消或超时 THEN THE System SHALL 恢复对应货道的虚拟库存
10. WHEN 货道虚拟库存为零 THEN THE System SHALL 标记该货道产品为不可购买状态
11. WHEN 货道实际库存为零但虚拟库存不为零 THEN THE System SHALL 触发库存异常警告
12. THE System SHALL 支持查询货道的库存差异（虚拟库存 - 实际库存）
13. THE System SHALL 支持管理员手动调整货道的虚拟库存和实际库存
14. WHEN 用户按设备 ID 筛选产品 THEN THE System SHALL 返回该设备所有货道的产品列表（按货道编号排序）
15. WHEN 用户按分类筛选产品 THEN THE System SHALL 返回该分类下的所有产品
16. WHEN 用户请求特定产品详情 THEN THE System SHALL 返回产品的完整信息（包括描述、营养信息）
17. WHEN 管理员更新产品价格 THEN THE System SHALL 验证价格为正数并更新记录
18. THE System SHALL 记录所有库存变更日志（时间、类型、数量、操作人）

### Requirement 5: 货道数据模型

**User Story:** 作为系统，我需要一个完整的货道数据模型，以支持设备、产品和库存的精确管理

#### Acceptance Criteria

1. THE System SHALL 创建 device_channels 表存储货道信息
2. WHEN 存储货道记录 THEN THE System SHALL 包含以下字段：id、device_id、channel_number、product_id、virtual_stock、actual_stock、max_capacity、status、created_at、updated_at
3. THE System SHALL 为每个货道的 channel_number 字段设置范围为 1-53
4. THE System SHALL 为 device_id 和 channel_number 创建唯一复合索引
5. THE System SHALL 设置 max_capacity 默认值为 4
6. THE System SHALL 验证 virtual_stock 和 actual_stock 不超过 max_capacity
7. THE System SHALL 支持货道状态：正常（normal）、故障（fault）、维护中（maintenance）、已禁用（disabled）
8. WHEN 货道状态为故障或已禁用 THEN THE System SHALL 不允许用户购买该货道产品
9. THE System SHALL 支持货道与产品的多对一关系（一个货道对应一个产品，一个产品可以在多个货道）
10. THE System SHALL 支持货道 product_id 为空（表示该货道未分配产品）
11. WHEN 查询设备货道 THEN THE System SHALL 关联返回产品信息（名称、价格、图片等）
12. THE System SHALL 为 device_id、product_id、status 字段创建索引

### Requirement 6: 订单管理与取货流程

**User Story:** 作为用户，我希望能够创建订单、支付并扫码取货，以便完成完整的购买流程

#### Acceptance Criteria

1. WHEN 用户提交有效的订单信息（设备 ID、货道编号、产品 ID、数量）THEN THE System SHALL 验证虚拟库存并创建订单
2. WHEN 用户提交订单时货道虚拟库存不足 THEN THE System SHALL 拒绝订单并返回库存不足错误
3. WHEN 订单创建成功 THEN THE System SHALL 扣减对应货道的虚拟库存并返回订单详情和支付信息
4. WHEN 订单创建成功 THEN THE System SHALL 生成唯一的取货二维码（包含订单 ID 和验证码）
5. WHEN 用户请求订单列表 THEN THE System SHALL 返回该用户的所有订单（按时间倒序）
6. WHEN 用户请求特定订单详情 THEN THE System SHALL 验证订单所有权并返回完整订单信息（包括取货二维码）
7. WHEN 订单支付完成 THEN THE System SHALL 更新订单状态为已支付并记录支付时间
8. WHEN 设备扫描取货二维码 THEN THE System SHALL 验证订单状态为已支付且未取货
9. WHEN 取货二维码验证成功 THEN THE System SHALL 返回货道编号和出货指令
10. WHEN 设备确认出货成功 THEN THE System SHALL 扣减对应货道的实际库存并更新订单状态为已完成
11. WHEN 设备出货失败 THEN THE System SHALL 记录失败原因并保持订单状态为已支付（允许重试）
12. WHEN 订单超过 30 分钟未支付 THEN THE System SHALL 自动取消订单并恢复虚拟库存
13. WHEN 订单已支付但超过 24 小时未取货 THEN THE System SHALL 标记订单为超时并恢复虚拟库存
14. WHEN 用户请求取消未支付订单 THEN THE System SHALL 取消订单并恢复虚拟库存
15. THE System SHALL 记录订单的所有状态变更（待支付、已支付、已完成、已取消、已超时）
16. THE System SHALL 支持订单包含多个货道的产品（购物车功能）
17. WHEN 订单包含多个产品 THEN THE System SHALL 验证所有货道的虚拟库存并原子性扣减

### Requirement 7: 支付系统

**User Story:** 作为用户，我希望能够使用微信或支付宝支付订单，以便完成购买

#### Acceptance Criteria

1. WHEN 用户选择微信支付方式 THEN THE System SHALL 生成微信支付二维码并返回支付参数
2. WHEN 用户选择支付宝支付方式 THEN THE System SHALL 生成支付宝支付二维码并返回支付参数
3. WHEN 支付平台发送支付成功回调 THEN THE System SHALL 验证签名并更新订单支付状态
4. WHEN 支付回调签名验证失败 THEN THE System SHALL 拒绝回调并记录安全日志
5. WHEN 用户查询订单支付状态 THEN THE System SHALL 返回当前支付状态（待支付、已支付、已取消、已退款）
6. WHEN 订单需要退款 THEN THE System SHALL 调用支付平台退款接口并更新订单状态
7. WHEN 支付平台返回退款成功 THEN THE System SHALL 更新订单状态为已退款并恢复库存

### Requirement 8: 媒体文件管理

**User Story:** 作为管理员，我希望能够上传和管理产品图片和其他媒体文件，以便展示产品信息并支持多种尺寸和格式转换

#### Acceptance Criteria

1. WHEN 管理员上传图片文件（JPEG、PNG、WebP、GIF）THEN THE System SHALL 验证文件类型和大小并存储文件
2. WHEN 上传的文件大小超过限制（10MB）THEN THE System SHALL 拒绝上传并返回文件过大错误
3. WHEN 上传的文件类型不支持 THEN THE System SHALL 拒绝上传并返回文件类型错误
4. WHEN 文件上传成功 THEN THE System SHALL 创建 Media 记录并返回文件的访问 URL
5. WHEN 上传图片文件 THEN THE System SHALL 自动生成多个尺寸的缩略图（thumb、medium、large）
6. WHEN 请求特定尺寸的图片 THEN THE System SHALL 返回对应尺寸的图片 URL
7. WHEN 删除媒体文件 THEN THE System SHALL 同时删除所有关联的缩略图和转换版本
8. WHEN 模型关联媒体文件 THEN THE System SHALL 支持多态关联（Product、User 等模型）
9. WHEN 查询模型的媒体文件 THEN THE System SHALL 返回该模型关联的所有媒体文件列表
10. WHEN 媒体文件关联到集合 THEN THE System SHALL 支持按集合名称组织媒体文件（如 images、documents）
11. THE System SHALL 在数据库中记录媒体文件的元数据（文件名、大小、MIME 类型、尺寸）
12. THE System SHALL 支持为媒体文件添加自定义属性（alt 文本、标题、描述）
13. WHEN 系统配置为使用 MinIO 存储 THEN THE System SHALL 将文件上传到 MinIO 服务器
14. WHEN 系统配置为使用 S3 存储 THEN THE System SHALL 将文件上传到 AWS S3
15. WHEN 系统配置为使用本地存储 THEN THE System SHALL 将文件保存到本地文件系统

### Requirement 9: 媒体库数据模型

**User Story:** 作为系统，我需要一个灵活的媒体库数据模型，以支持多种实体类型关联媒体文件

#### Acceptance Criteria

1. THE System SHALL 创建 media 表存储媒体文件信息
2. WHEN 存储媒体记录 THEN THE System SHALL 包含以下字段：id、model_type、model_id、collection_name、name、file_name、mime_type、disk、size、manipulations、custom_properties、order_column、created_at、updated_at
3. WHEN 媒体文件关联到模型 THEN THE System SHALL 使用 model_type 和 model_id 实现多态关联
4. WHEN 媒体文件属于集合 THEN THE System SHALL 使用 collection_name 字段标识集合（如 images、avatars、documents）
5. WHEN 生成缩略图或转换 THEN THE System SHALL 在 manipulations 字段存储转换配置（JSON 格式）
6. WHEN 添加自定义属性 THEN THE System SHALL 在 custom_properties 字段存储自定义数据（JSON 格式）
7. WHEN 查询模型的媒体文件 THEN THE System SHALL 按 order_column 字段排序返回
8. THE System SHALL 为 model_type、model_id、collection_name 字段创建复合索引
9. THE System SHALL 支持软删除媒体记录
10. WHEN 删除关联模型 THEN THE System SHALL 级联删除该模型的所有媒体文件

### Requirement 10: 数据验证

**User Story:** 作为系统，我需要验证所有输入数据，以确保数据完整性和安全性

#### Acceptance Criteria

1. WHEN 接收到 API 请求 THEN THE System SHALL 验证所有必填字段是否存在
2. WHEN 邮箱字段存在 THEN THE System SHALL 验证邮箱格式是否正确
3. WHEN 密码字段存在 THEN THE System SHALL 验证密码长度至少为 8 个字符
4. WHEN 价格字段存在 THEN THE System SHALL 验证价格为正数
5. WHEN 数量字段存在 THEN THE System SHALL 验证数量为正整数
6. WHEN 验证失败 THEN THE System SHALL 返回 422 错误和详细的验证错误信息
7. WHEN 验证成功 THEN THE System SHALL 继续处理请求

### Requirement 11: 自定义异常系统

**User Story:** 作为开发者，我希望系统提供自定义异常类型，以便更精确地处理和表达不同类型的错误

#### Acceptance Criteria

1. THE System SHALL 提供基础 ApiException 异常类型
2. THE System SHALL 提供以下预定义异常类型：ValidationException、AuthenticationException、AuthorizationException、NotFoundException、BadRequestException、ConflictException、ServerException
3. WHEN 创建 ApiException THEN THE System SHALL 包含以下属性：message（错误消息）、code（HTTP 状态码）、errors（详细错误信息）
4. WHEN 抛出 ValidationException THEN THE System SHALL 自动设置 HTTP 状态码为 422
5. WHEN 抛出 AuthenticationException THEN THE System SHALL 自动设置 HTTP 状态码为 401
6. WHEN 抛出 AuthorizationException THEN THE System SHALL 自动设置 HTTP 状态码为 403
7. WHEN 抛出 NotFoundException THEN THE System SHALL 自动设置 HTTP 状态码为 404
8. WHEN 抛出 BadRequestException THEN THE System SHALL 自动设置 HTTP 状态码为 400
9. WHEN 抛出 ConflictException THEN THE System SHALL 自动设置 HTTP 状态码为 409
10. WHEN 抛出 ServerException THEN THE System SHALL 自动设置 HTTP 状态码为 500
11. THE System SHALL 支持自定义异常消息和错误详情
12. THE System SHALL 支持链式调用设置异常属性（如 WithMessage、WithErrors、WithCode）
13. WHEN 捕获 ApiException THEN THE System SHALL 自动转换为统一的 JSON 响应格式

### Requirement 12: 错误处理

**User Story:** 作为开发者，我希望系统能够统一处理错误，以便前端能够正确显示错误信息，并在开发环境中获得详细的调试信息

#### Acceptance Criteria

1. WHEN 发生 ApiException THEN THE System SHALL 使用异常定义的状态码和消息返回响应
2. WHEN 发生未捕获的异常 THEN THE System SHALL 返回 500 状态码并记录错误日志
3. WHEN 发生数据库错误 THEN THE System SHALL 包装为 ServerException 并返回 500 状态码
4. THE System SHALL 以统一的 JSON 格式返回所有错误响应（包含 success、code、message、errors 字段）
5. WHEN 环境变量 APP_DEBUG 为 true THEN THE System SHALL 在错误响应中添加 debugger 对象
6. WHEN 包含 debugger 对象 THEN THE System SHALL 在其中包含 exception 字段（异常类名）
7. WHEN 包含 debugger 对象 THEN THE System SHALL 在其中包含 trace 数组（堆栈跟踪信息）
8. WHEN 包含 debugger 对象 THEN THE System SHALL 在其中包含性能信息（请求处理时间、内存使用）
9. WHEN 环境变量 APP_DEBUG 为 false THEN THE System SHALL 不在响应中包含 debugger 对象和敏感信息
10. THE System SHALL 使用全局错误处理中间件捕获所有异常
11. THE System SHALL 记录所有错误到日志系统

### Requirement 13: 调试模式响应格式

**User Story:** 作为开发者，我希望在开发环境中能够获得详细的调试信息，以便快速定位和解决问题

#### Acceptance Criteria

1. WHEN 环境变量 APP_DEBUG 为 true 且发生错误 THEN THE System SHALL 在响应中包含 debugger 对象
2. WHEN debugger 对象存在 THEN THE System SHALL 包含 exception 字段（字符串类型，异常类名）
3. WHEN debugger 对象存在 THEN THE System SHALL 包含 trace 字段（数组类型，堆栈跟踪信息）
4. WHEN debugger 对象存在 THEN THE System SHALL 包含 file 字段（字符串类型，发生错误的文件路径）
5. WHEN debugger 对象存在 THEN THE System SHALL 包含 line 字段（整数类型，发生错误的行号）
6. WHEN debugger 对象存在 THEN THE System SHALL 包含 request_time 字段（浮点数类型，请求处理时间，单位毫秒）
7. WHEN debugger 对象存在 THEN THE System SHALL 包含 memory_usage 字段（字符串类型，内存使用量）
8. WHEN debugger 对象存在 THEN THE System SHALL 包含 query_count 字段（整数类型，数据库查询次数）
9. WHEN 环境变量 APP_DEBUG 为 false THEN THE System SHALL 不在任何响应中包含 debugger 对象
10. THE System SHALL 确保生产环境（APP_DEBUG=false）不泄露敏感的调试信息

### Requirement 14: API 响应格式

**User Story:** 作为前端开发者，我希望所有 API 响应格式统一，以便简化前端处理逻辑

#### Acceptance Criteria

1. THE System SHALL 以 JSON 格式返回所有 API 响应
2. WHEN API 请求成功 THEN THE System SHALL 返回包含 success（true）、code、message、data 字段的响应
3. WHEN API 请求失败 THEN THE System SHALL 返回包含 success（false）、code、message、errors 字段的响应
4. WHEN 返回单个资源或操作结果 THEN THE System SHALL 将数据放在 data 字段中
5. WHEN 返回列表数据 THEN THE System SHALL 在 data 字段中包含 items 数组和 meta 分页信息对象
6. WHEN 返回分页数据 THEN THE System SHALL 在 meta 对象中包含 current_page、per_page、last_page、has_more、total、from、to 字段
7. WHEN 返回验证错误 THEN THE System SHALL 在 errors 对象中按字段名组织错误信息数组
8. WHEN 返回时间字段 THEN THE System SHALL 使用 ISO 8601 格式（YYYY-MM-DDTHH:mm:ssZ）
9. WHEN 返回金额字段 THEN THE System SHALL 使用分为单位的整数
10. THE System SHALL 在响应头中设置正确的 Content-Type 为 application/json
11. THE System SHALL 确保 success 字段为布尔类型，code 字段为整数类型

### Requirement 15: 数据库迁移和填充

**User Story:** 作为开发者，我希望能够使用迁移系统管理数据库结构，以便版本控制和团队协作

#### Acceptance Criteria

1. WHEN 执行迁移命令 THEN THE System SHALL 按顺序执行所有未运行的迁移文件
2. WHEN 执行回滚命令 THEN THE System SHALL 回滚最后一批迁移
3. WHEN 迁移文件包含 Up 方法 THEN THE System SHALL 在迁移时执行该方法
4. WHEN 迁移文件包含 Down 方法 THEN THE System SHALL 在回滚时执行该方法
5. WHEN 执行数据填充命令 THEN THE System SHALL 运行所有 Seeder 文件
6. THE System SHALL 记录已执行的迁移到 migrations 表
7. THE System SHALL 支持创建表、修改表结构、添加索引等数据库操作

### Requirement 16: 缓存系统

**User Story:** 作为系统，我需要缓存热点数据，以提高 API 响应速度

#### Acceptance Criteria

1. WHEN 查询城市列表 THEN THE System SHALL 优先从缓存读取数据
2. WHEN 缓存中不存在数据 THEN THE System SHALL 从数据库查询并写入缓存
3. WHEN 城市信息更新 THEN THE System SHALL 清除相关缓存
4. WHEN 设备信息更新 THEN THE System SHALL 清除相关缓存
5. WHEN 产品信息更新 THEN THE System SHALL 清除相关缓存
6. THE System SHALL 使用 Redis 作为缓存存储
7. THE System SHALL 为缓存数据设置合理的过期时间（默认 1 小时）

### Requirement 17: 日志记录

**User Story:** 作为运维人员，我希望系统能够记录详细日志，以便排查问题和监控系统

#### Acceptance Criteria

1. THE System SHALL 记录所有 API 请求日志（包含请求路径、方法、参数、响应时间）
2. THE System SHALL 记录所有错误日志（包含错误类型、堆栈信息、上下文）
3. THE System SHALL 记录所有数据库查询日志（在开发环境）
4. THE System SHALL 记录所有支付相关操作日志
5. THE System SHALL 使用结构化日志格式（JSON）
6. THE System SHALL 支持日志级别配置（DEBUG、INFO、WARN、ERROR）
7. THE System SHALL 将日志输出到文件并支持日志轮转

### Requirement 18: 性能要求

**User Story:** 作为用户，我希望 API 响应速度快，以获得良好的使用体验

#### Acceptance Criteria

1. WHEN 查询列表接口 THEN THE System SHALL 在 200ms 内返回响应（90% 的请求）
2. WHEN 查询详情接口 THEN THE System SHALL 在 100ms 内返回响应（90% 的请求）
3. WHEN 创建订单接口 THEN THE System SHALL 在 500ms 内返回响应（90% 的请求）
4. THE System SHALL 使用数据库连接池管理连接
5. THE System SHALL 为常用查询添加数据库索引
6. THE System SHALL 使用 Redis 缓存热点数据
7. THE System SHALL 支持 API 响应压缩（gzip）

### Requirement 19: 任务调度系统

**User Story:** 作为系统管理员，我希望能够定时执行任务，以便自动化处理定期任务（如订单清理、数据统计等）

#### Acceptance Criteria

1. THE System SHALL 提供类似 Laravel Schedule 的任务调度功能
2. WHEN 定义调度任务 THEN THE System SHALL 支持 Cron 表达式配置执行时间
3. WHEN 定义调度任务 THEN THE System SHALL 支持链式方法配置（每分钟、每小时、每天、每周、每月）
4. THE System SHALL 支持以下预定义调度频率：EveryMinute、EveryFiveMinutes、EveryTenMinutes、EveryThirtyMinutes、Hourly、Daily、Weekly、Monthly
5. WHEN 启动调度器 THEN THE System SHALL 持续运行并按配置的时间执行任务
6. WHEN 执行调度任务 THEN THE System SHALL 记录任务执行日志（开始时间、结束时间、执行结果）
7. WHEN 调度任务执行失败 THEN THE System SHALL 记录错误日志并继续执行其他任务
8. THE System SHALL 支持任务互斥锁，防止同一任务重复执行
9. THE System SHALL 支持任务超时配置
10. THE System SHALL 提供命令行工具启动调度器（如 go run main.go schedule:run）
11. WHEN 配置调度任务 THEN THE System SHALL 支持在代码中注册任务（类似 Laravel Kernel）
12. THE System SHALL 支持以下常见调度任务：清理过期订单、生成日报、清理临时文件、数据备份

### Requirement 20: 队列系统

**User Story:** 作为开发者，我希望能够使用队列处理异步任务，以提高系统响应速度和处理耗时操作

#### Acceptance Criteria

1. THE System SHALL 提供类似 Laravel Queue 的队列系统
2. THE System SHALL 使用 Redis 作为队列后端存储
3. WHEN 分发任务到队列 THEN THE System SHALL 将任务序列化并存储到 Redis
4. WHEN 启动队列工作进程 THEN THE System SHALL 提供命令行工具（如 go run main.go queue:work）
5. WHEN 队列工作进程运行 THEN THE System SHALL 持续从队列中获取并执行任务
6. THE System SHALL 支持多个队列（default、high、low 等优先级队列）
7. WHEN 分发任务 THEN THE System SHALL 支持指定队列名称和延迟时间
8. WHEN 任务执行失败 THEN THE System SHALL 支持自动重试机制（可配置重试次数）
9. WHEN 任务重试次数超过限制 THEN THE System SHALL 将任务移到失败队列
10. THE System SHALL 记录所有队列任务的执行日志
11. THE System SHALL 支持以下常见队列任务：发送邮件、发送通知、处理图片、生成报表、数据导入导出
12. THE System SHALL 支持配置队列工作进程的并发数
13. THE System SHALL 支持优雅关闭队列工作进程（处理完当前任务后退出）
14. THE System SHALL 提供查看队列状态的命令（如 go run main.go queue:status）

### Requirement 21: WebSocket 实时通信

**User Story:** 作为用户，我希望能够实时接收系统通知和订单状态更新，以获得更好的用户体验

#### Acceptance Criteria

1. THE System SHALL 提供类似 Laravel Reverb 的 WebSocket 服务
2. THE System SHALL 使用 WebSocket 协议实现客户端和服务器的双向通信
3. WHEN 客户端连接 WebSocket THEN THE System SHALL 验证客户端的认证 Token
4. WHEN 客户端认证失败 THEN THE System SHALL 拒绝 WebSocket 连接
5. THE System SHALL 支持频道（Channel）概念组织消息
6. THE System SHALL 支持公共频道（任何人都可以订阅）
7. THE System SHALL 支持私有频道（需要认证才能订阅）
8. THE System SHALL 支持存在频道（Presence Channel，显示在线用户列表）
9. WHEN 客户端订阅频道 THEN THE System SHALL 验证客户端是否有权限订阅该频道
10. WHEN 服务器广播消息 THEN THE System SHALL 将消息推送到订阅该频道的所有客户端
11. THE System SHALL 支持以下事件类型：订单状态更新、支付成功通知、设备状态变更、库存预警
12. WHEN 订单状态变更 THEN THE System SHALL 通过 WebSocket 实时通知用户
13. WHEN 支付成功 THEN THE System SHALL 通过 WebSocket 实时通知用户
14. THE System SHALL 记录 WebSocket 连接日志（连接、断开、订阅、取消订阅）
15. THE System SHALL 支持 WebSocket 连接心跳检测
16. THE System SHALL 支持自动重连机制（客户端）
17. THE System SHALL 提供命令行工具启动 WebSocket 服务器（如 go run main.go websocket:serve）
18. THE System SHALL 支持配置 WebSocket 服务器端口和路径

### Requirement 22: 技术选型要求

**User Story:** 作为开发团队，我希望使用 Golang 生态中最流行和成熟的开源包，以确保项目的稳定性、可维护性和社区支持

#### Acceptance Criteria

1. THE System SHALL 优先选择 GitHub Stars 数量高的开源包（通常 > 1000 stars）
2. THE System SHALL 优先选择活跃维护的开源包（最近 6 个月内有更新）
3. THE System SHALL 使用 Fiber v2 作为 Web 框架（高性能、Express 风格）
4. THE System SHALL 使用 GORM 作为 ORM 框架（最流行的 Go ORM）
5. THE System SHALL 使用 PostgreSQL 作为主数据库
6. THE System SHALL 使用 Redis 作为缓存和队列后端
7. THE System SHALL 使用 MinIO 作为对象存储服务
8. THE System SHALL 使用 Viper 进行配置管理（最流行的配置库）
9. THE System SHALL 使用 golang-jwt/jwt 进行 JWT 认证
10. THE System SHALL 使用 redis/go-redis 作为 Redis 客户端
11. THE System SHALL 使用 Asynq 作为队列系统
12. THE System SHALL 使用 go-playground/validator 进行数据验证
13. THE System SHALL 使用 Cobra 构建命令行工具
14. THE System SHALL 使用 Zap 作为日志库
15. THE System SHALL 使用 Swaggo 生成 Swagger API 文档
16. THE System SHALL 使用 golang.org/x/crypto/bcrypt 进行密码加密
17. THE System SHALL 使用 gorilla/websocket 实现 WebSocket
18. THE System SHALL 使用 go-co-op/gocron 实现任务调度
19. THE System SHALL 使用 disintegration/imaging 进行图片处理
20. THE System SHALL 使用 minio/minio-go 作为 MinIO 客户端
21. THE System SHALL 使用 spf13/cast 进行类型转换
22. THE System SHALL 避免使用不活跃或缺乏维护的第三方包
23. THE System SHALL 在选择第三方包时考虑许可证兼容性（优先 MIT、Apache 2.0）
24. THE System SHALL 定期更新依赖包以获取安全补丁和新特性

### Requirement 23: 安全要求

**User Story:** 作为系统管理员，我希望系统具有良好的安全性，以保护用户数据和系统安全

#### Acceptance Criteria

1. THE System SHALL 使用 bcrypt 算法加密存储用户密码
2. THE System SHALL 使用 HTTPS 传输敏感数据（生产环境）
3. THE System SHALL 验证所有 JWT Token 的签名和有效期
4. THE System SHALL 防止 SQL 注入攻击（使用 ORM 参数化查询）
5. THE System SHALL 防止 XSS 攻击（转义输出内容）
6. THE System SHALL 实现 API 请求频率限制（每分钟最多 60 次请求）
7. THE System SHALL 记录所有安全相关事件（登录失败、权限拒绝等）
8. THE System SHALL 验证 WebSocket 连接的认证 Token
9. THE System SHALL 防止 WebSocket 消息注入攻击
