-- 创建钱包状态枚举
CREATE TYPE wallet_status AS ENUM ('active', 'inactive', 'suspended', 'frozen');

-- 创建钱包表
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- 关联用户
    balance DECIMAL(20, 8) NOT NULL DEFAULT 500.00, -- TRU余额，默认500
    frozen_balance DECIMAL(20, 8) NOT NULL DEFAULT 0.00, -- 冻结余额
    status wallet_status NOT NULL DEFAULT 'active', -- 钱包状态
    is_withdrawal_enabled BOOLEAN NOT NULL DEFAULT false, -- 是否允许提现
    transaction_pin_hash VARCHAR(255), -- 交易密码哈希
    pin_attempts INTEGER NOT NULL DEFAULT 0, -- 密码错误尝试次数
    pin_locked_until TIMESTAMP WITH TIME ZONE, -- 密码锁定至
    max_pin_attempts INTEGER NOT NULL DEFAULT 5, -- 最大密码尝试次数
    last_transaction_at TIMESTAMP WITH TIME ZONE, -- 最后交易时间
    daily_withdrawal_limit DECIMAL(20, 8) DEFAULT 100000.00, -- 每日提现限额（TRU）
    daily_withdrawn_amount DECIMAL(20, 8) NOT NULL DEFAULT 0.00, -- 今日已提现金额
    last_withdrawal_reset TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- 上次重置提现限额时间
    withdrawal_count INTEGER NOT NULL DEFAULT 0, -- 提现次数统计
    total_deposited DECIMAL(20, 8) NOT NULL DEFAULT 500.00, -- 累计充值金额
    total_withdrawn DECIMAL(20, 8) NOT NULL DEFAULT 0.00, -- 累计提现金额
    notes TEXT, -- 备注信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束检查
    CONSTRAINT check_balance_non_negative CHECK (balance >= 0),
    CONSTRAINT check_frozen_balance_non_negative CHECK (frozen_balance >= 0),
    CONSTRAINT check_daily_withdrawn_non_negative CHECK (daily_withdrawn_amount >= 0),
    CONSTRAINT check_total_amounts_non_negative CHECK (total_deposited >= 0 AND total_withdrawn >= 0),
    CONSTRAINT check_pin_attempts_non_negative CHECK (pin_attempts >= 0),
    CONSTRAINT check_withdrawal_count_non_negative CHECK (withdrawal_count >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_status ON wallets(status);
CREATE INDEX IF NOT EXISTS idx_wallets_is_withdrawal_enabled ON wallets(is_withdrawal_enabled);
CREATE INDEX IF NOT EXISTS idx_wallets_balance ON wallets(balance);
CREATE INDEX IF NOT EXISTS idx_wallets_last_transaction_at ON wallets(last_transaction_at);
CREATE INDEX IF NOT EXISTS idx_wallets_created_at ON wallets(created_at);

-- 为所有现有用户创建钱包
INSERT INTO wallets (user_id, balance, total_deposited)
SELECT id, 500.00, 500.00 
FROM users 
WHERE NOT EXISTS (SELECT 1 FROM wallets WHERE wallets.user_id = users.id);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_wallets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_wallets_updated_at();

-- 创建自动为新用户创建钱包的触发器
CREATE OR REPLACE FUNCTION create_wallet_for_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO wallets (user_id, balance, total_deposited)
    VALUES (NEW.id, 500.00, 500.00);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_create_wallet_for_new_user
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION create_wallet_for_new_user();