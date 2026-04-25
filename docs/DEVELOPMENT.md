# 本地开发与规范

## 常用命令

- 启动开发：`make dev`
- 直接运行：`make run`
- 单元测试：`make test`
- 全量检查：`make check`（fmt + vet + lint + test）

## CLI

CLI 入口为 `cmd/cli/main.go`，常用：

- 列出路由：`make routes`
- 生成 JWT 密钥：`make jwt`
- 运行队列 worker：`go run ./cmd/cli queue:work`

## 迁移（Atlas / SQLite）

- 生成迁移：`make atlas-diff-sqlite NAME=xxx`
- 应用迁移：`make atlas-apply-sqlite`

说明：SQLite 以 `database/schema.lt.hcl` 为 schema-as-code；迁移目录固定为 `database/migrations/atlas/sqlite`。

## 测试约束

- 单测默认不依赖外部服务。
- 需要外部依赖（Redis/Garage/Postgres 等）的测试必须通过显式环境变量开启。
