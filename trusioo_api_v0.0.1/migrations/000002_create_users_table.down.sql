-- 删除用户表的索引
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email_verified;
DROP INDEX IF EXISTS idx_users_created_at;

-- 删除用户表
DROP TABLE IF EXISTS users;