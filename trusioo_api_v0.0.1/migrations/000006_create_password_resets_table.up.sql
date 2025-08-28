-- 创建密码重置表
CREATE TABLE IF NOT EXISTS password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    user_type VARCHAR(50) NOT NULL, -- admin, user
    token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    used BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets(email);
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
CREATE INDEX IF NOT EXISTS idx_password_resets_user_type ON password_resets(user_type);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_resets_used ON password_resets(used);
CREATE INDEX IF NOT EXISTS idx_password_resets_created_at ON password_resets(created_at);

-- 创建复合索引用于查找有效的重置令牌
CREATE INDEX IF NOT EXISTS idx_password_resets_valid 
ON password_resets(email, user_type, used, expires_at);

-- 创建清理过期记录的函数
CREATE OR REPLACE FUNCTION cleanup_expired_password_resets()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM password_resets 
    WHERE expires_at < NOW() - INTERVAL '30 days'; -- 保留30天的过期记录用于审计
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;