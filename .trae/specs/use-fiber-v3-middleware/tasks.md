# Tasks

- [X] Task 1: 盘点并确定替换范围

  - [X] SubTask 1.1: 列出 setup.go 实际启用的中间件与顺序
  - [X] SubTask 1.2: 标注可用 Fiber v3 官方 middleware 替换的自定义实现
- [x] Task 2: 替换 request id 为官方 requestid middleware

  - [x] SubTask 2.1: 移除自定义 RequestID() 的挂载点
  - [x] SubTask 2.2: 接入 `github.com/gofiber/fiber/v3/middleware/requestid`
  - [x] SubTask 2.3: 确保 `X-Request-ID` 生成/透传/回写行为符合 Spec
- [x] Task 3: 替换访问日志为官方 logger middleware

  - [x] SubTask 3.1: 移除内联 zap access log 中间件
  - [x] SubTask 3.2: 接入 `github.com/gofiber/fiber/v3/middleware/logger` 并配置格式包含 request id、latency、status、method、path
  - [x] SubTask 3.3: 校验 middleware 顺序（favicon 在 logger 之前，requestid 在 logger 之前）
- [x] Task 4: 测试与回归

  - [x] SubTask 4.1: 调整/新增测试以覆盖 request id 行为与日志包含 request id
  - [x] SubTask 4.2: 运行 `go test ./...` 并修复回归

# Task Dependencies

- Task 3 depends on Task 2
- Task 4 depends on Task 2, Task 3
