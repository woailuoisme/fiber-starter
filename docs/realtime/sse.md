# SSE 协议说明

SSE 是 `pkg/realtime` 的只读传输方式，和 WebSocket 共享频道、广播、Redis bus 与服务端 `Dispatch`。它面向浏览器原生 `EventSource`，默认不支持客户端向服务端发送事件。

## 连接

- **URL**: `GET /sse/app/<app_key>`
- **Content-Type**: `text/event-stream`
- **订阅参数**: `channels=public-news,private-orders.1`
- **保活**: 服务端按 `heartbeat_interval` 发送 `: ping` comment。

公共频道可直接订阅：

```text
/sse/app/lfiber?channels=public-news
```

私有或 Presence 频道需要附带 Pusher 风格签名参数：

```text
/sse/app/lfiber?channels=private-orders.1&socket_id=sse-1&auth=lfiber:signature
```

多私有频道可使用 `auths` JSON 映射；Presence 频道如需签名载荷，可使用 `channel_data` 或 `channel_data_by_channel`。

## 事件帧

SSE 的 `event:` 字段使用业务事件名，`data:` 字段是完整 JSON envelope：

```text
id: 8a1f...
event: orders.updated
data: {"channel":"private-orders.1","event":"orders.updated","data":{"id":1},"socket_id":"socket-1"}
```

## 语义边界

- SSE 不计入 Presence 成员，不触发 `member_added` / `member_removed`。
- SSE 不支持离线补偿和 `Last-Event-ID` 回放；断线期间的事件按在线即达语义处理。
- 慢客户端发送队列满时会被断开，以保护服务端内存和广播延迟。
