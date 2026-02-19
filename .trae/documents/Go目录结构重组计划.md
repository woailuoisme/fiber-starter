# Go 目录结构重组计划（从 Laravel 风格到 Go Pro）

## 目标

- 将当前 `app/`、`bootstrap/`、`command/` 等“Laravel 风格目录”迁移为更 idiomatic 的 Go 工程结构：`cmd/` + `internal/`（必要时 `pkg/`）
- 保持功能不变：HTTP 服务与 CLI 仍可正常启动，测试可通过
- 降低未来维护成本：分层清晰（transport / service / domain / platform），避免跨层依赖与循环引用

## 现状摘要（关键点）

- 入口已相对标准：`cmd/server/main.go`、`cmd/cli/main.go`
- 启动链路集中：`bootstrap/app.go`
- HTTP 分层集中：`app/http/*` + `app/routers/*`
- CLI 分层集中：`command/*`（Cobra）
- 已存在 `internal/db/gen/...`（生成代码），可作为 `internal/` 体系的一部分保留

## 目标目录结构（建议）

> 原则：`cmd/` 只放薄入口；核心实现尽量放 `internal/`；只有明确需要被外部复用的库才放 `pkg/`。

- `cmd/`
  - `server/`：HTTP 服务入口（薄 main）
  - `cli/`：CLI 入口（薄 main）
- `internal/`
  - `app/`：应用装配（DI 容器、启动/退出、wire-up）
    - `httpserver/`：Fiber App 构建、监听、优雅退出
    - `providers/`：依赖注册（原 `app/providers`）
  - `transport/`
    - `http/`
      - `controller/`（原 `app/http/controllers`）
      - `middleware/`（原 `app/http/middleware`）
      - `router/`（原 `app/routers`）
      - `request/`（原 `app/http/requests`）
      - `response/`（原 `app/http/resources`）
  - `service/`：业务服务（原 `app/services`）
  - `domain/`
    - `model/`（原 `app/models`）
    - `enum/`（原 `app/enums`）
  - `platform/`：横切能力（日志、缓存、错误、工具函数等）
    - `errors/`（原 `app/apierrors` + `app/exceptions`）
    - `helpers/`（原 `app/helpers`，后续可再按能力拆包）
  - `i18n/`：i18n 逻辑（原 `app/i18n`）
  - `scheduler/`：调度（原 `app/schedule`）
  - `cli/`：cobra 命令实现（原 `command`）
  - `db/`：数据库连接/seed（原 `database/connection.go`、`database/seeders/*.go`）
    - `gen/`：保留现有生成代码目录（当前已存在）
- `configs/`：配置文件（yaml）保持不变（如需变更查找路径，再单独调整）
- `database/migrations/`：迁移文件保持不变（它们是数据资产，不强制归入 Go 包）
- `docs/`、`scripts/`、`tests/`：保持不变（仅按 import/路径变化做必要调整）
- `lang/`：可选改名为 `locales/`（若要与 Go 社区惯例对齐）；否则先保持不动

## 迁移映射（从旧到新）

- `bootstrap/` → `internal/app/httpserver/`
  - `bootstrap/app.go` → `internal/app/httpserver/app.go`（或 `server.go`）
- `app/providers/` → `internal/app/providers/`
- `app/http/controllers/` → `internal/transport/http/controller/`
- `app/http/middleware/` → `internal/transport/http/middleware/`
- `app/http/requests/` → `internal/transport/http/request/`
- `app/http/resources/` → `internal/transport/http/response/`
- `app/routers/` → `internal/transport/http/router/`
- `app/services/` → `internal/service/`
- `app/models/` → `internal/domain/model/`
- `app/enums/` → `internal/domain/enum/`
- `app/helpers/` → `internal/platform/helpers/`
- `app/apierrors/` + `app/exceptions/` → `internal/platform/errors/`
- `app/i18n/` → `internal/i18n/`
- `app/schedule/` → `internal/scheduler/`
- `command/` → `internal/cli/`
- `database/connection.go` → `internal/db/connection.go`
- `database/seeders/*.go` → `internal/db/seeder/`（目录名可定为 `seed` 或 `seeder`，以现有包名为准）
- `config/` → `internal/config/`（推荐）

## 执行顺序（确保每一步可编译）

1. **建立目标目录骨架（只新增目录，不先删旧目录）**
2. **先搬迁“叶子包”（依赖少的包）并更新 import**
   - `internal/domain/*`、`internal/platform/*`、`internal/config`、`internal/db`（不含生成代码）
3. **迁移 HTTP 层**
   - controllers/middleware/router/request/response，修复相互引用与 import 路径
4. **迁移 DI 与启动链路**
   - providers → internal/app/providers
   - bootstrap/app.go → internal/app/httpserver，并让 `cmd/server/main.go` 只做调用
5. **迁移 CLI**
   - `command/*` → `internal/cli/*`，让 `cmd/cli/main.go` 只做调用
6. **清理旧目录**
   - 确认无引用后删除 `app/`、`bootstrap/`、`command/` 旧目录（及 `database/connection.go` 等已迁走文件）
7. **验证与修复**
   - 单元测试、lint、启动验证（server 与 cli）

## 关键改动点（需要同步调整）

- 所有 Go 文件的 `package` 名称与目录一致性（迁移时优先保持原 package 名不变，避免一次性改动过大；后续再做包名精炼）
- 所有 `import "fiber-starter/app/..."`、`import "fiber-starter/bootstrap"`、`import "fiber-starter/command"` 等需要批量替换为 `fiber-starter/internal/...`
- `docs` 的 swagger 引用（`_ "fiber-starter/docs"`）保持不变
- 测试包（`tests/*.go`）的 import 路径需要同步更新

## 验收标准

- `go test ./...` 全绿
- `cmd/server` 能启动并正常响应现有健康检查/业务路由
- `cmd/cli` 能执行现有命令（至少：routes、migrate、schedule、queue_work 等）
- `golangci-lint`（若当前项目已在 CI/脚本中启用）无新增关键问题

## 风险与规避

- **循环依赖风险**：迁移时严格分层（transport 只能依赖 service/domain/platform；service 依赖 domain/platform；platform 不依赖上层）
- **大范围 import 替换风险**：按“叶子包 → 上层包”的顺序迁移，确保每一步都能编译/测试
- **路径/资源查找风险**（configs、lang、migrations 等）：优先保持目录不动，仅调整 Go 包组织；若代码中使用相对路径读取资源，再按需要统一为基于项目根或可配置路径
