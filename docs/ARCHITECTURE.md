# 架构与分层约定

本项目目标是：**高可用基线**（优雅退出、健康检查、结构化日志）+ **结构化工程**（单一错误出口、依赖边界清晰、可测试、可演进）。

## 分层概览

- `cmd/server/` + `bootstrap/`：应用启动与生命周期管理（创建 DI 容器、Fiber app、优雅退出、资源关闭）。
- `app/providers/`：依赖注入（dig）注册入口；把配置、数据库连接、缓存等基础设施注入给上层。
- `app/http/routers/`：路由组装（只负责“把 handler 挂到路径上”）。
- `app/http/controllers/`：HTTP 适配层（解析输入、调用 service、返回结果/错误）。
- `app/services/`：业务服务（组合 DB/缓存/队列/存储等依赖完成业务逻辑）。
- `database/`：数据库连接、schema-as-code、迁移与 seed。
- `command/`：CLI 命令集合（migrate / routes / queue:work / jwt:generate 等）。
- `app/http/middleware/`：HTTP 中间件（request_id、日志、安全、鉴权、错误出口等）。

## 关键工程契约（必须遵守）

### 单一错误出口

- HTTP 层所有错误统一交给 `app/http/middleware/error_handler.go` 输出响应。
- controller/middleware **返回 `error` 即可**，不要在各处拼装多套 JSON 错误格式。

### 依赖注入优先，避免关键全局

- 业务代码避免直接依赖全局单例（配置、DB、缓存等）；优先从 dig 注入。
- 路由层需要的鉴权中间件由启动层（bootstrap）通过注入的配置/缓存构造后传入。

### JWT 与登出语义

- 登录生成 access/refresh token；refresh token 存缓存。
- 登出会把 access token 写入缓存黑名单，鉴权中间件会检查黑名单使 token 立即失效。

## 数据库与迁移（SQLite 优先）

- SQLite：以 Atlas 作为唯一迁移来源
  - schema：`database/schema.lt.hcl`
  - migrations：`database/migrations/atlas/sqlite`
- Postgres：当前阶段不作为重构重点（如需启用，再补齐迁移/校验策略）。
