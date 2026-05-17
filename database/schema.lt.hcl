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

table "auth_otps" {
  schema = schema.main

  column "id" {
    type           = integer
    null           = false
    auto_increment = true
  }

  column "email" {
    type = text
    null = false
  }

  column "purpose" {
    type = text
    null = false
  }

  column "code_hash" {
    type = text
    null = false
  }

  column "expires_at" {
    type = datetime
    null = false
  }

  column "sent_at" {
    type = datetime
    null = false
  }

  column "attempts" {
    type    = integer
    null    = false
    default = 0
  }

  column "max_attempts" {
    type    = integer
    null    = false
    default = 5
  }

  column "consumed_at" {
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

  primary_key {
    columns = [column.id]
  }

  index "idx_auth_otps_email_purpose_created_at" {
    columns = [column.email, column.purpose, column.created_at]
  }

  index "idx_auth_otps_expires_at" {
    columns = [column.expires_at]
  }

  check "ck_auth_otps_purpose" {
    expr = "purpose IN ('signup','password_reset')"
  }

  check "ck_auth_otps_attempts" {
    expr = "attempts >= 0"
  }

  check "ck_auth_otps_max_attempts" {
    expr = "max_attempts > 0"
  }
}
