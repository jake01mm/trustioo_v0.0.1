-- 创建用户会话表
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    user_type VARCHAR(50) NOT NULL, -- admin, user
    refresh_token_id UUID, -- 关联refresh_tokens表
    ip_address INET NOT NULL,
    user_agent TEXT,
    device_info JSONB,
    location_info JSONB, -- IP地理位置信息
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 外键约束
    CONSTRAINT fk_user_sessions_refresh_token FOREIGN KEY (refresh_token_id) REFERENCES refresh_tokens(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_type ON user_sessions(user_type);
CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh_token_id ON user_sessions(refresh_token_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_last_activity ON user_sessions(last_activity);
CREATE INDEX IF NOT EXISTS idx_user_sessions_created_at ON user_sessions(created_at);

-- 创建复合索引用于查找用户的活跃会话
CREATE INDEX IF NOT EXISTS idx_user_sessions_active 
ON user_sessions(user_id, user_type, is_active, expires_at);

-- 创建清理过期会话的函数
CREATE OR REPLACE FUNCTION cleanup_expired_user_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- 标记过期会话为非活跃状态
    UPDATE user_sessions 
    SET is_active = false, last_activity = NOW()
    WHERE is_active = true AND expires_at < NOW();
    
    -- 删除7天前的非活跃会话
    DELETE FROM user_sessions 
    WHERE is_active = false AND last_activity < NOW() - INTERVAL '7 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;