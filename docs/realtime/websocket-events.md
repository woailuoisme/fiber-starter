# WebSocket Event 列表

本文档列出了系统支持的 WebSocket 内部事件及通用业务事件。

## 1. 内部协议事件 (Internal Protocol Events)

这些事件由客户端或服务端用于控制连接状态和订阅。

| 事件名 | 方向 | 说明 |
| :--- | :--- | :--- |
| `pusher:connection_established` | S -> C | 连接建立成功，返回 `socket_id`。 |
| `pusher:subscribe` | C -> S | 客户端请求订阅通道。 |
| `pusher:unsubscribe` | C -> S | 客户端请求取消订阅。 |
| `pusher:ping` | Both | 心跳检测。 |
| `pusher:pong` | Both | 心跳响应。 |
| `pusher:error` | S -> C | 发生错误（如授权失败、频率限制等）。 |
| `pusher_internal:subscription_succeeded` | S -> C | 订阅成功确认。如果是成员通道，会携带当前成员列表。 |

## 2. 成员通道事件 (Presence Events)

仅在 `presence-` 前缀的通道中触发。

| 事件名 | 方向 | 说明 |
| :--- | :--- | :--- |
| `pusher_internal:member_added` | S -> C | 有新成员加入通道。 |
| `pusher_internal:member_removed` | S -> C | 有成员离开通道（或断开连接）。 |

## 3. 示例载荷 (Payload Examples)

### 3.1 `pusher:connection_established`

```json
{
  "event": "pusher:connection_established",
  "data": {
    "socket_id": "84c8a29a-50f9-4673-8a07-882d2946c108",
    "activity_timeout": 30
  }
}
```

### 3.2 `pusher:subscribe` (私有通道)

```json
{
  "event": "pusher:subscribe",
  "data": {
    "channel": "private-orders.1",
    "auth": "app_key:signature"
  }
}
```

### 3.3 `pusher_internal:subscription_succeeded` (成员通道)

```json
{
  "event": "pusher_internal:subscription_succeeded",
  "channel": "presence-chat.room.1",
  "data": {
    "presence": {
      "count": 2,
      "members": [
        { "user_id": "1", "user_info": { "name": "Alice" } },
        { "user_id": "2", "user_info": { "name": "Bob" } }
      ]
    }
  }
}
```

### 3.4 业务广播 (Client-Side Trigger)

如果你服务端允许，客户端可以向已订阅的私有通道发送广播：

```json
{
  "event": "client-typing",
  "channel": "private-chat.1",
  "data": { "is_typing": true }
}
```

*注：客户端事件必须以 `client-` 为前缀。*
