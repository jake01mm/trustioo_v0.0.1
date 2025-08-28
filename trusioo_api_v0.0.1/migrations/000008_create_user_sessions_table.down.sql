-- 删除清理函数
DROP FUNCTION IF EXISTS cleanup_expired_user_sessions();

-- 删除用户会话表
DROP TABLE IF EXISTS user_sessions;