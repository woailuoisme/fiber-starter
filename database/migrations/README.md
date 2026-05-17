# 数据库迁移

本目录用于版本化数据库迁移，结构尽量贴近 Laravel 的 `database/migrations` 习惯。

强约束：

- `database/schema.pg.hcl` 和 `database/schema.lt.hcl` 是 schema source of truth。
- Bun 模型是应用侧 schema 入口，Atlas 根据 Bun 模型生成或比对迁移。
- 所有 DDL 迁移必须由 Atlas 基于 Bun 模型或 HCL 生成，不能手写维护为主。

- PostgreSQL：`database/migrations/postgres`
- SQLite：`database/migrations/sqlite`

推荐工作流：

1. 修改 Bun 模型：
   - `app/Models/user.go`
   - `app/Models/auth_otp.go`
2. 必要时同步维护 schema 文件：
   - PostgreSQL：`database/schema.pg.hcl`
   - SQLite：`database/schema.lt.hcl`
3. 生成迁移：
   - `atlas migrate diff <name> --env postgres`
   - `atlas migrate diff <name> --env sqlite`
4. 应用迁移：
   - `atlas migrate apply --env postgres`
   - `atlas migrate apply --env sqlite`
5. 刷新校验：
   - `atlas migrate hash --env postgres`
   - `atlas migrate hash --env sqlite`

当前项目保留按数据库类型分目录的迁移方式，方便同时支持 PostgreSQL 和 SQLite。
