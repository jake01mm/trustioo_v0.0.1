-- 删除清理函数
DROP FUNCTION IF EXISTS cleanup_expired_password_resets();

-- 删除密码重置表
DROP TABLE IF EXISTS password_resets;