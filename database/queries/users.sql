-- name: CountUsers :one
SELECT COUNT(1) FROM users WHERE deleted_at IS NULL;

-- name: CountUsersBySearch :one
SELECT COUNT(1)
FROM users
WHERE deleted_at IS NULL
  AND (LOWER(name) LIKE LOWER(sqlc.arg(pattern)) OR LOWER(email) LIKE LOWER(sqlc.arg(pattern)));

-- name: UserExistsByEmail :one
SELECT EXISTS(
  SELECT 1 FROM users
  WHERE email = sqlc.arg(email) AND deleted_at IS NULL
);

-- name: GetUserByEmail :one
SELECT id, name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at
FROM users
WHERE email = sqlc.arg(email) AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByID :one
SELECT id, name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at
FROM users
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
LIMIT 1;

-- name: ListUsers :many
SELECT id, name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
  AND (
    sqlc.arg(pattern) = ''
    OR LOWER(name) LIKE LOWER(sqlc.arg(pattern))
    OR LOWER(email) LIKE LOWER(sqlc.arg(pattern))
  )
ORDER BY created_at DESC
LIMIT sqlc.arg(limit)
OFFSET sqlc.arg(offset);

-- name: CreateUser :one
INSERT INTO users (
  name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at
) VALUES (
  sqlc.arg(name), sqlc.arg(email), sqlc.arg(password), sqlc.narg(avatar), sqlc.narg(phone),
  sqlc.arg(status), sqlc.narg(email_verified_at), sqlc.arg(created_at), sqlc.arg(updated_at)
) RETURNING id, name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at;

-- name: UpdateUser :exec
UPDATE users
SET
  name = sqlc.arg(name),
  email = sqlc.arg(email),
  password = sqlc.arg(password),
  avatar = sqlc.narg(avatar),
  phone = sqlc.narg(phone),
  status = sqlc.arg(status),
  email_verified_at = sqlc.narg(email_verified_at),
  updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: UpdatePassword :exec
UPDATE users
SET password = sqlc.arg(password), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: ResetPasswordByEmail :exec
UPDATE users
SET password = sqlc.arg(password), updated_at = sqlc.arg(updated_at)
WHERE email = sqlc.arg(email) AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = sqlc.arg(deleted_at), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: DeleteAllUsers :exec
DELETE FROM users;
