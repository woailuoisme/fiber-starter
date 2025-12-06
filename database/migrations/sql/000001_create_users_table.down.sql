-- 删除触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- 删除用户表
DROP TABLE IF EXISTS users;
