# Atlas 迁移目录

本目录用于 Atlas 的版本化迁移（versioned migrations）。

- PostgreSQL：`database/migrations/atlas/postgres`
- SQLite：`database/migrations/atlas/sqlite`

推荐工作流：

1. 修改 schema 源文件：
   - PostgreSQL：`database/schema.pg.hcl`
   - SQLite：`database/schema.lt.hcl`
2. 生成迁移：
   - `atlas migrate diff <name> --env postgres`
   - `atlas migrate diff <name> --env sqlite`
3. 应用迁移：
   - `atlas migrate apply --env postgres`
   - `atlas migrate apply --env sqlite`
