-- 删除清理函数
DROP FUNCTION IF EXISTS cleanup_old_login_logs();

-- 删除登录日志表
DROP TABLE IF EXISTS login_logs;