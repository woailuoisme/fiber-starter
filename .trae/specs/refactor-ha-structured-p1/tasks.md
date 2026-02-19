# Tasks
- [x] Task 1: 统一 request_id 读取与日志字段
  - [x] SubTask 1.1: 定义 request_id 的 header、locals key 与字段名规范
  - [x] SubTask 1.2: 对齐 access log 与 error handler 的 request_id 读取方式
  - [x] SubTask 1.3: 让错误日志以结构化字段输出 request_id（不拼接 message）
  - [x] SubTask 1.4: 添加/更新单元测试覆盖 request_id 生成、沿用与回写行为

- [x] Task 2: 去全局状态并收敛到 DI（增量迁移）
  - [x] SubTask 2.1: 列出业务层与命令层对 `config.GlobalConfig` 的直接引用点并制定迁移顺序
  - [x] SubTask 2.2: 对 CLI 命令（migrate/seed 等）改为从容器构造依赖，不直接读取全局
  - [x] SubTask 2.3: 对 `database.DB` 的直接引用点增加过渡适配与替换路径（优先 command/seed）
  - [x] SubTask 2.4: 添加静态扫描/测试约束，防止新增关键全局引用回流

- [x] Task 3: 新增队列 worker 运维命令 `queue:work`
  - [x] SubTask 3.1: 新增 cobra 子命令并接入现有容器初始化流程
  - [x] SubTask 3.2: 在命令中启动 asynq server（复用 `QueueService.StartWorker()` 或等价入口）
  - [x] SubTask 3.3: 纳入优雅退出：signal 监听 + stop/shutdown + 资源释放
  - [x] SubTask 3.4: 添加命令级测试（至少覆盖命令注册与启动路径不 panic）

- [x] Task 4: 回归验证与文档补齐（最小化）
  - [x] SubTask 4.1: 运行 `go test ./...` 并修复回归
  - [x] SubTask 4.2: 本地验证 /health 与 /ready 行为不回退
  - [x] SubTask 4.3: 补齐 README/运行手册中 worker 运行方式（单独进程）

# Task Dependencies
- Task 3 depends on Task 2 (worker 命令需要稳定的容器/依赖构造)
- Task 4 depends on Task 1, Task 2, Task 3
