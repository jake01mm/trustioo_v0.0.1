-- 创建第三方OAuth令牌表
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    user_type VARCHAR(50) NOT NULL, -- admin, user
    provider VARCHAR(50) NOT NULL, -- google, github, wechat, weibo, qq等
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    provider_name VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    scope TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    provider_data JSONB, -- 存储来自提供商的额外用户信息
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保每个用户每个提供商只有一条记录
    UNIQUE(user_id, provider)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user_id ON oauth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user_type ON oauth_tokens(user_type);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_provider ON oauth_tokens(provider);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_provider_user_id ON oauth_tokens(provider, provider_user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_provider_email ON oauth_tokens(provider_email);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_is_active ON oauth_tokens(is_active);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_expires_at ON oauth_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_created_at ON oauth_tokens(created_at);

-- 创建复合索引用于快速查找用户的OAuth绑定
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user_provider 
ON oauth_tokens(user_id, user_type, provider, is_active);

-- 创建用于防止重复绑定的唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_tokens_provider_unique 
ON oauth_tokens(provider, provider_user_id) 
WHERE is_active = true;

-- 创建清理过期令牌的函数
CREATE OR REPLACE FUNCTION cleanup_expired_oauth_tokens()
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    -- 将过期的令牌标记为非活跃状态
    UPDATE oauth_tokens 
    SET is_active = false, updated_at = NOW()
    WHERE is_active = true 
    AND expires_at IS NOT NULL 
    AND expires_at < NOW();
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;