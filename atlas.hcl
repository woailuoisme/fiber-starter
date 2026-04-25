variable "pg_url" {
  type    = string
  default = getenv("DATABASE_URL")
}

env "postgres" {
  schema {
    src = "file://database/schema.pg.hcl"
  }

  migration {
    dir = "file://database/migrations/postgres"
  }

  url = var.pg_url

  dev = "postgres://postgres:password@host.docker.internal:5432/postgres?search_path=public&sslmode=disable"
}

env "sqlite" {
  schema {
    src = "file://database/schema.lt.hcl"
  }

  migration {
    dir = "file://database/migrations/sqlite"
  }

  url = "sqlite://database/lunchbox_vending.sqlite?_fk=1"

  dev = "sqlite://file?mode=memory&_fk=1"
}
