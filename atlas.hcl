# 从环境变量中获取核心物理数据库 URL，默认由本地 .env 自动注入
variable "pg_url" {
  type    = string
  default = getenv("DATABASE_URL")
}

# 挂载外部 Bun 驱动，用于自动从 Go 代码中加载并解析 Postgres 实体模型 Schema
data "external_schema" "bun_postgres" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path",
    "./internal",
    "--dialect",
    "postgres",
  ]
}

# 挂载外部 Bun 驱动，用于自动从 Go 代码中加载并解析 SQLite 实体模型 Schema
data "external_schema" "bun_sqlite" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path",
    "./internal",
    "--dialect",
    "sqlite",
  ]
}

# Postgres 迁移与执行环境配置
env "postgres" {
  schema {
    src = data.external_schema.bun_postgres.url
  }

  migration {
    # 迁移脚本输出路径
    dir = "file://database/migrations/postgres"
  }

  # 目标数据库连接
  url = var.pg_url

  # 开发辅助的容器临时数据库（用于 Schema Diff 比对计算）
  dev = "postgres://postgres:password@host.docker.internal:5432/postgres?search_path=public&sslmode=disable"
}

# SQLite 迁移与执行环境配置
env "sqlite" {
  schema {
    src = data.external_schema.bun_sqlite.url
  }

  migration {
    # 迁移脚本输出路径
    dir = "file://database/migrations/sqlite"
  }

  # 本地开发 SQLite 单机库路径
  url = "sqlite://database/lunchbox_vending.sqlite?_fk=1"

  # 开发辅助的内存临时数据库（用于 Schema Diff 比对计算）
  dev = "sqlite://file?mode=memory&_fk=1"
}

