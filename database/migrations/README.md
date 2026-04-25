# 数据库迁移

本目录用于版本化数据库迁移，结构尽量贴近 Laravel 的 `database/migrations` 习惯。

- PostgreSQL：`database/migrations/postgres`
- SQLite：`database/migrations/sqlite`

推荐工作流：

1. 修改数据库模式定义：
   - PostgreSQL：`database/schema.pg.hcl`
   - SQLite：`database/schema.lt.hcl`
2. 生成迁移：
   - `atlas migrate diff <name> --env postgres`
   - `atlas migrate diff <name> --env sqlite`
3. 应用迁移：
   - `atlas migrate apply --env postgres`
   - `atlas migrate apply --env sqlite`

当前项目保留按数据库类型分目录的迁移方式，方便同时支持 PostgreSQL 和 SQLite。
