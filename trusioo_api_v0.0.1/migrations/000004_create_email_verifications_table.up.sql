-- 创建邮箱验证表
CREATE TABLE IF NOT EXISTS email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    user_type VARCHAR(50) NOT NULL, -- admin, user
    type VARCHAR(50) NOT NULL, -- login, password_reset, email_change, profile_change, wallet_activation, withdrawal, payment, account_security, deactivation
    verification_code VARCHAR(10) NOT NULL, -- 6位数字验证码
    token VARCHAR(255) UNIQUE, -- 备用验证令牌
    attempts INTEGER NOT NULL DEFAULT 0, -- 尝试次数
    max_attempts INTEGER NOT NULL DEFAULT 3,
    verified BOOLEAN NOT NULL DEFAULT false,
    ip_address INET,
    reference_id UUID, -- 可选：关联具体的操作ID
    metadata JSONB, -- 可选：存储额外的验证相关数据
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_email_verifications_email ON email_verifications(email);
CREATE INDEX IF NOT EXISTS idx_email_verifications_type ON email_verifications(type);
CREATE INDEX IF NOT EXISTS idx_email_verifications_email_type ON email_verifications(email, type);
CREATE INDEX IF NOT EXISTS idx_email_verifications_code ON email_verifications(verification_code);
CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at);
CREATE INDEX IF NOT EXISTS idx_email_verifications_verified ON email_verifications(verified);
CREATE INDEX IF NOT EXISTS idx_email_verifications_reference_id ON email_verifications(reference_id);

-- 创建复合唯一索引，防止同一邮箱同一类型有多个未过期的验证码
CREATE UNIQUE INDEX IF NOT EXISTS idx_email_verifications_unique_active 
ON email_verifications(email, type, user_type) 
WHERE verified = false AND expires_at > NOW();

-- 创建清理过期记录的函数
CREATE OR REPLACE FUNCTION cleanup_expired_email_verifications()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM email_verifications 
    WHERE expires_at < NOW() - INTERVAL '7 days'; -- 保留7天的过期记录用于审计
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;