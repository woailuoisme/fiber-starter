variable "pg_url" {
  type    = string
  default = getenv("DATABASE_URL")
}

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

env "postgres" {
  schema {
    src = data.external_schema.bun_postgres.url
  }

  migration {
    dir = "file://database/migrations/postgres"
  }

  url = var.pg_url

  dev = "postgres://postgres:password@host.docker.internal:5432/postgres?search_path=public&sslmode=disable"
}

env "sqlite" {
  schema {
    src = data.external_schema.bun_sqlite.url
  }

  migration {
    dir = "file://database/migrations/sqlite"
  }

  url = "sqlite://database/lunchbox_vending.sqlite?_fk=1"

  dev = "sqlite://file?mode=memory&_fk=1"
}
