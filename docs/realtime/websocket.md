# WebSocket 协议说明

本文档描述了本系统的 WebSocket 协议实现。本协议采用了类 Pusher 的消息格式，以便于与现有的 Pusher 客户端（如 `pusher-js`）或兼容库集成。

## 1. 连接与认证

### 1.1 建立连接

WebSocket 端点通常位于 `/ws` 或 `/realtime`。

- **URL**: `ws://<domain>/ws`
- **协议**: 推荐使用标准 WebSocket。

连接成功后，服务端会自动发送 `pusher:connection_established` 事件。

### 1.2 私有通道认证

私有通道（`private-` 前缀）和成员通道（`presence-` 前缀）需要进行服务端认证。

- **Payload**:

  ```json
  {
    "socket_id": "xxx-xxx-xxx",
    "channel_name": "private-user.1"
  }
  ```

- **响应**:

  ```json
  {
    "auth": "app_key:signature",
    "channel_data": "{...}" // 仅成员通道包含
  }
  ```

## 2. 消息格式

所有消息均采用 JSON 格式传输：

```json
{
  "event": "event_name",
  "channel": "channel_name",
  "data": { ... },
  "socket_id": "optional_origin_socket_id"
}
```

- **event**: 事件名称。
- **channel**: 通道名称（订阅/取消订阅时必须，公共广播时可选）。
- **data**: 事件载荷（JSON 对象或字符串）。
- **socket_id**: 发送者的 Socket ID（用于排除发送者自身广播）。

## 3. 通道类型

| 类型 | 前缀 | 说明 |
| :--- | :--- | :--- |
| 公共通道 | 无 | 任何人均可订阅。 |
| 私有通道 | `private-` | 需要认证，仅授权用户可订阅。 |
| 成员通道 | `presence-` | 基于私有通道，增加在线成员追踪功能。 |

## 4. 客户端交互流程

1. **连接**: 客户端连接 WebSocket。
2. **握手**: 接收 `pusher:connection_established` 获取 `socket_id`。
3. **认证** (可选): 如果是私有通道，调用 API 获取 `auth`。
4. **订阅**: 发送 `pusher:subscribe`。
5. **接收**: 监听自定义业务事件或内部状态变更。
6. **保活**: 定期接收 `pusher:ping` 并响应 `pusher:pong`（或者相反）。
