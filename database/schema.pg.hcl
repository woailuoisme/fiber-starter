schema "public" {}

table "users" {
  schema = schema.public

  column "id" {
    type           = bigint
    null           = false
    identity {
      generated = "BY DEFAULT"
    }
  }

  column "name" {
    type = varchar(255)
    null = false
  }

  column "email" {
    type = varchar(255)
    null = false
  }

  column "password" {
    type = varchar(255)
    null = false
  }

  column "avatar" {
    type = varchar(500)
    null = true
  }

  column "phone" {
    type = varchar(20)
    null = true
  }

  column "status" {
    type    = varchar(20)
    null    = false
    default = "active"
  }

  column "email_verified_at" {
    type = timestamptz
    null = true
  }

  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("CURRENT_TIMESTAMP")
  }

  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("CURRENT_TIMESTAMP")
  }

  column "deleted_at" {
    type = timestamptz
    null = true
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_users_email" {
    unique  = true
    columns = [column.email]
  }

  index "idx_users_status" {
    columns = [column.status]
  }

  index "idx_users_deleted_at" {
    columns = [column.deleted_at]
  }

  check "ck_users_status" {
    expr = "status IN ('active','inactive','banned')"
  }
}
