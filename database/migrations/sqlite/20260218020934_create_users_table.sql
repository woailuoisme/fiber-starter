-- Create "users" table
CREATE TABLE `users` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `email` text NOT NULL,
  `password` text NOT NULL,
  `avatar` text NULL,
  `phone` text NULL,
  `status` text NOT NULL DEFAULT 'active',
  `email_verified_at` datetime NULL,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `deleted_at` datetime NULL,
  CONSTRAINT `ck_users_status` CHECK (status IN ('active','inactive','banned'))
);
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX `idx_users_email` ON `users` (`email`);
-- Create index "idx_users_status" to table: "users"
CREATE INDEX `idx_users_status` ON `users` (`status`);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX `idx_users_deleted_at` ON `users` (`deleted_at`);
