-- 删除管理员表的索引
DROP INDEX IF EXISTS idx_admins_email;
DROP INDEX IF EXISTS idx_admins_active;
DROP INDEX IF EXISTS idx_admins_created_at;

-- 删除管理员表
DROP TABLE IF EXISTS admins;