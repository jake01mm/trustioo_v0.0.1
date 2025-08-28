-- 创建刷新令牌表
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_id VARCHAR(255) UNIQUE NOT NULL, -- JWT中的jti字段
    user_id UUID NOT NULL,
    user_type VARCHAR(50) NOT NULL, -- admin, user
    device_info JSONB, -- 设备信息(user_agent, ip等)
    ip_address INET,
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_id ON refresh_tokens(token_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_type ON refresh_tokens(user_type);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_created_at ON refresh_tokens(created_at);

-- 创建复合索引用于查找用户的有效令牌
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_valid 
ON refresh_tokens(user_id, user_type, is_revoked, expires_at);