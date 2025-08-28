-- 删除清理函数
DROP FUNCTION IF EXISTS cleanup_expired_oauth_tokens();

-- 删除第三方OAuth令牌表
DROP TABLE IF EXISTS oauth_tokens;