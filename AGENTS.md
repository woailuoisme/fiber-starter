# Repository Guidelines

## 项目结构与模块组织

这是一个 Go 1.26 + Fiber v3 的后端项目。HTTP API 入口在 `cmd/server`，Cobra 命令入口在 `cmd/cli`。应用代码采用 Laravel 风格的顶层包组织：`app/Http` 放控制器、中间件、请求、资源和 HTTP 用例服务；`app/Console` 放命令和调度器内核；`app/Models`、`app/Enums`、`app/Exceptions`、`app/I18n`、`app/Providers`、`app/Services`、`app/Support` 分别承载模型、枚举、异常、国际化、依赖注入、基础设施服务和公共支持层。配置在 `config/`，数据库结构在 `database/`，其中 `database/factories`、`database/migrations`、`database/seeders` 分别对应工厂、迁移和种子；路由在 `routes/`，OpenAPI 输出在 `docs/`，语言文件在 `lang/`，静态资源在 `public/`，测试在 `tests/`。构建与覆盖率输出默认值由 `.buildconfig` 控制，并通过 `make config` 和 `scripts/config.sh` 查看。

## 构建、测试与开发命令

- `make init` 安装工具、同步依赖，并在缺少时从 `.env.example` 创建 `.env`。
- `make dev` 优先使用 Air 启动 API，否则回退到 `go run ./cmd/server`。
- `make run` 直接启动服务。
- `make build` 构建 `build/fiber-starter`；`make build-cli` 构建 `build/fiber-starter-cli`。
- `make build-prod` 构建去符号的 Linux 服务端二进制。
- `make config` 打印 `.buildconfig` 生效后的构建/输出配置。
- `make test` 运行 `go test -v ./...`。
- `make coverage` 在 `coverage/` 下生成 HTML 和 profile 报告。
- `make check` 运行格式化、vet、lint 和测试。
- `make docs` 重新生成 `docs/` 下的 OpenAPI/Swagger 文件供 Scalar 使用。
- `make atlas-diff NAME=<name>` 和 `make atlas-apply` 用于管理迁移；需要时可设置 `ENV=sqlite`。

## 编码风格与命名约定

遵循 Go 的惯用写法：使用 `gofmt` 的制表符风格、短包名、显式错误处理，以及 `%w` 的上下文包装。除非需要重新生成，否则不要改动 `database/`、`docs/` 或其他生成输出目录里的文件。优先通过 `app/Providers` 做依赖装配，不要依赖包级全局变量。HTTP 面向的用例放在 `app/Http/Services`，基础设施服务放在 `app/Services`。提交前运行 `make fmt` 和 `make lint`；lint 包含 `govet`、`staticcheck`、`errcheck`、`revive`、`gosec`。

## 测试规范

测试使用 Go 标准 `testing` 包和 Fiber 的 `app.Test` + `httptest`。测试名建议使用 `TestFeature_Behavior` 或 `TestFeatureDoesNotRegress`，跨包回归测试放在 `tests/`。中间件、控制器、CLI 命令、数据库行为以及安全敏感改动都要补充针对性测试。

## 提交与 PR 规范

当前历史习惯使用简短的小写提交，例如 `fix`；请保持提交简洁、聚焦单一任务。更清晰的提交名更好，例如 `fix request id logging` 或 `add sqlite migration`。PR 需要包含摘要、测试结果（例如 `make check`）、关联 issue，以及涉及 HTTP 行为或文档变更时的截图或 API 示例。涉及迁移、新环境变量和生成代码时要在说明里写清楚。

## 安全与配置提示

不要硬编码密钥；使用 `.env`，并保持 `.env.example` 同步更新。校验 HTTP 和 CLI 输入，避免在日志里打印凭证或 token；发布前运行 `make lint`，因为已经启用了 `gosec`。推送前检查 `git diff`，尤其是生成文档、迁移文件和 vendored 依赖变更。
