schema "main" {}

table "users" {
  schema = schema.main

  column "id" {
    type           = integer
    null           = false
    auto_increment = true
  }

  column "name" {
    type = text
    null = false
  }

  column "email" {
    type = text
    null = false
  }

  column "password" {
    type = text
    null = false
  }

  column "avatar" {
    type = text
    null = true
  }

  column "phone" {
    type = text
    null = true
  }

  column "status" {
    type    = text
    null    = false
    default = "active"
  }

  column "email_verified_at" {
    type = datetime
    null = true
  }

  column "created_at" {
    type    = datetime
    null    = false
    default = sql("CURRENT_TIMESTAMP")
  }

  column "updated_at" {
    type    = datetime
    null    = false
    default = sql("CURRENT_TIMESTAMP")
  }

  column "deleted_at" {
    type = datetime
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
