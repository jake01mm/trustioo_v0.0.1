-- 创建登录日志表
CREATE TABLE IF NOT EXISTS login_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    email VARCHAR(255) NOT NULL,
    user_type VARCHAR(50) NOT NULL, -- admin, user
    login_status VARCHAR(50) NOT NULL, -- success, failed, blocked, suspicious
    failure_reason VARCHAR(255), -- 失败原因
    ip_address INET NOT NULL,
    user_agent TEXT,
    device_info JSONB, -- 详细设备信息
    location_info JSONB, -- IP地理位置信息
    session_id VARCHAR(255), -- 会话ID
    risk_score INTEGER DEFAULT 0, -- 风险评分 0-100
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_login_logs_user_id ON login_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_login_logs_email ON login_logs(email);
CREATE INDEX IF NOT EXISTS idx_login_logs_user_type ON login_logs(user_type);
CREATE INDEX IF NOT EXISTS idx_login_logs_ip_address ON login_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_logs_login_status ON login_logs(login_status);
CREATE INDEX IF NOT EXISTS idx_login_logs_created_at ON login_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_login_logs_session_id ON login_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_login_logs_risk_score ON login_logs(risk_score);

-- 创建复合索引用于安全分析
CREATE INDEX IF NOT EXISTS idx_login_logs_security 
ON login_logs(email, ip_address, login_status, created_at);

-- 创建复合索引用于用户登录历史查询
CREATE INDEX IF NOT EXISTS idx_login_logs_user_history 
ON login_logs(user_id, user_type, created_at DESC);

-- 创建清理旧日志的函数
CREATE OR REPLACE FUNCTION cleanup_old_login_logs()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM login_logs 
    WHERE created_at < NOW() - INTERVAL '1 year'; -- 保留1年的登录日志
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;