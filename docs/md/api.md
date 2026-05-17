# Fiber Starter 生产级 API 规范指南

欢迎使用本项目生成的交互式 API 文档。

## 🚀 快速上手

本项目基于 **Fiber v3** 开发，遵循 RESTful 架构风格。

### 认证流程

大多数接口需要 **Bearer Token** 认证。

1. 调用 `/api/v1/auth/sign-in` 获取令牌。
2. 在请求头中包含：`Authorization: Bearer <your_token>`。

## 📦 统一响应格式

所有接口均返回统一的 JSON 结构，便于前端处理：

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `success` | bool | 操作是否成功 |
| `code` | int | 业务或 HTTP 状态码 |
| `message` | string | 提示信息 |
| `data` | object | 具体的业务数据内容 |
| `exception` | object | 调试模式下的异常堆栈（仅 Debug 模式可见） |

## 🛠 开发规范

- **分页参数**：使用 `page` 和 `limit`。
- **错误处理**：返回非 2xx 状态码时，请检查 `errors` 字段。
- **时间格式**：统一采用 RFC3339 格式。

---

> 本文档由 Antigravity 自动优化生成。
