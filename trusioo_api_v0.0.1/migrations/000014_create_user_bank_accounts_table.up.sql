-- 创建银行账户状态枚举
CREATE TYPE bank_account_status AS ENUM ('active', 'inactive', 'suspended', 'pending_verification');

-- 创建用户银行账户表
CREATE TABLE IF NOT EXISTS user_bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- 关联用户
    bank_id UUID NOT NULL REFERENCES banks(id), -- 关联银行
    account_number VARCHAR(50) NOT NULL, -- 银行账户号码
    account_name VARCHAR(100) NOT NULL, -- 账户姓名
    account_type VARCHAR(20) DEFAULT 'savings', -- 账户类型：savings, current, checking
    sort_code VARCHAR(20), -- 分行代码
    iban VARCHAR(50), -- 国际银行账户号码（如适用）
    bic_code VARCHAR(15), -- 银行识别代码（如适用）
    status bank_account_status NOT NULL DEFAULT 'pending_verification', -- 账户状态
    is_default BOOLEAN NOT NULL DEFAULT false, -- 是否为默认账户
    is_verified BOOLEAN NOT NULL DEFAULT false, -- 是否已验证
    verification_method VARCHAR(20), -- 验证方式：manual, automatic, document
    verified_at TIMESTAMP WITH TIME ZONE, -- 验证时间
    verified_by UUID, -- 验证人（管理员ID）
    verification_notes TEXT, -- 验证备注
    usage_count INTEGER NOT NULL DEFAULT 0, -- 使用次数统计
    last_used_at TIMESTAMP WITH TIME ZONE, -- 最后使用时间
    notes TEXT, -- 备注信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保同一用户在同一银行只能有一个相同账户号码
    CONSTRAINT unique_user_bank_account UNIQUE (user_id, bank_id, account_number)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_user_id ON user_bank_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_bank_id ON user_bank_accounts(bank_id);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_account_number ON user_bank_accounts(account_number);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_status ON user_bank_accounts(status);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_is_default ON user_bank_accounts(is_default);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_is_verified ON user_bank_accounts(is_verified);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_verified_at ON user_bank_accounts(verified_at);
CREATE INDEX IF NOT EXISTS idx_user_bank_accounts_created_at ON user_bank_accounts(created_at);

-- 创建部分索引：每个用户只能有一个默认账户
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_default_bank_account 
ON user_bank_accounts(user_id) 
WHERE is_default = true;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_user_bank_accounts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_user_bank_accounts_updated_at
    BEFORE UPDATE ON user_bank_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_user_bank_accounts_updated_at();

-- 创建设置默认账户时自动取消其他默认账户的触发器
CREATE OR REPLACE FUNCTION ensure_single_default_account()
RETURNS TRIGGER AS $$
BEGIN
    -- 如果新记录被设置为默认账户
    IF NEW.is_default = true THEN
        -- 取消该用户的其他默认账户
        UPDATE user_bank_accounts 
        SET is_default = false, updated_at = NOW()
        WHERE user_id = NEW.user_id 
        AND id != NEW.id 
        AND is_default = true;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_ensure_single_default_account
    BEFORE INSERT OR UPDATE ON user_bank_accounts
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_default_account();