-- 删除清理函数
DROP FUNCTION IF EXISTS cleanup_expired_email_verifications();

-- 删除邮箱验证表
DROP TABLE IF EXISTS email_verifications;