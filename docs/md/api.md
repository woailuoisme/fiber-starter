# Fiber Go 后端服务 API 规范与接入指南

欢迎使用本项目自动生成的交互式 API 接口文档。本项目基于 Go 1.26.3 和 Fiber v3 构建，采用 Feature-First（特性优先）的自治架构设计。为了保证前后端协作的高效性与系统稳定性，请遵循以下规范进行接口对接与调用。

## 1. 快速上手

### 1.1 服务端点与文档

- **API 基础前缀**：`/api/v1`
- **在线文档地址**：`/docs`（提供基于 Scalar UI 的交互式接口调试界面，定义文件为 `/openapi.json`）

### 1.2 身份验证流程

大多数核心业务接口均受 JWT 保护，采用标准 Bearer Token 机制：

1. 访问认证接口 `/api/v1/auth/sign-in` 或 `/api/v1/auth/register` 获取有效的 JWT 访问令牌。
2. 后续请求必须在 HTTP 请求头中携带：

   ```http
   Authorization: Bearer <your_access_token>
   ```

3. 令牌失效或缺失时，系统将统一返回 `401 Unauthorized` 状态码。

---

## 2. 统一响应格式 (Uniform JSON Structure)

系统所有 HTTP 接口均返回标准化的 JSON 数据结构。这保证了客户端在解析数据结构时具有强确定性。

### 2.1 基础结构体定义

| 字段名 | 数据类型 | 说明 |
| :--- | :--- | :--- |
| `success` | boolean | 指示操作在逻辑上是否成功 |
| `code` | integer | 业务状态码或 HTTP 状态码 |
| `message` | string | 对客户端的友好提示信息或错误描述 |
| `data` | object / array | 具体的业务数据内容，操作失败或无返回数据时为具体空实体或空数组 |
| `exception` | object / null | 仅在系统开启 Debug 模式下返回的异常堆栈信息，生产环境隐藏此字段 |

### 2.2 响应示例

#### 成功响应示例

```json
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1001,
    "username": "seaside",
    "email": "user@example.com"
  }
}
```

#### 空集合序列化规范 (CRITICAL)

根据项目开发规范，在 Marshaling 序列化流程中，**任何集合/切片类型的数据绝不返回 `nil`**。空数据集必须初始化为具体的空切片 `[]`，从而彻底杜绝 JavaScript、Kotlin、Swift 等强类型客户端在解析 JSON 时发生空指针异常 (Null Pointer Crash)。

```json
{
  "success": true,
  "code": 200,
  "message": "查询成功",
  "data": []
}
```

#### 失败响应与参数校验失败示例

当请求参数未通过结构体验证（如邮箱格式错误、必填项缺失）时，系统会返回 `422 Unprocessable Entity` 错误，并在 `errors` 或 `message` 字段中详细说明：

```json
{
  "success": false,
  "code": 422,
  "message": "请求参数校验未通过",
  "errors": {
    "email": "邮箱格式不正确",
    "password": "密码长度不能小于 6 位"
  }
}
```

---

## 3. 开发与接入规范

### 3.1 协议规范

- **传输安全**：所有生产环境通信必须采用 HTTPS。
- **时间格式**：时间字段统一采用 ISO 8601 / RFC 3339 格式，如 `2026-05-18T19:16:41+08:00`。
- **编码格式**：请求体 and 响应体统一采用 `application/json; charset=utf-8`。

### 3.2 统一分页模型

对于涉及列表查询的接口，请求统一采用 Query 参数：

- `page`：当前页码（从 1 开始）
- `limit`：每页数据量限制

响应体中的 `data` 节点通常包装以下结构：

```json
{
  "success": true,
  "code": 200,
  "message": "查询成功",
  "data": {
    "items": [],
    "total": 0,
    "current_page": 1,
    "per_page": 15,
    "last_page": 1
  }
}
```

### 3.3 异常与 HTTP 状态码映射

项目严格遵守标准 HTTP 状态码语义：

| 状态码 | 含义 | 说明 |
| :--- | :--- | :--- |
| `200 OK` | 请求成功 | 请求成功执行，且有返回数据。 |
| `201 Created` | 创建成功 | 资源创建成功。 |
| `204 No Content` | 无内容 | 请求成功执行，无响应内容。 |
| `400 Bad Request` | 请求错误 | 客户端请求格式错误或语义错误。 |
| `401 Unauthorized` | 未授权 | 未提供身份凭证或凭证已过期。 |
| `403 Forbidden` | 拒绝访问 | 权限不足，拒绝访问该资源。 |
| `404 Not Found` | 未找到 | 请求的资源不存在。 |
| `405 Method Not Allowed` | 方法不允许 | 请求的 HTTP 方法不被该路由支持。 |
| `422 Unprocessable Entity` | 实体无法处理 | 请求数据格式正确，但业务或验证规则校验失败。 |
| `429 Too Many Requests` | 请求过于频繁 | 触发接口速率限制（Rate Limiting）。 |
| `500 Internal Server Error` | 服务器内部错误 | 服务端内部异常，请联系系统管理员。 |
