-- 创建交易类型枚举
CREATE TYPE transaction_type AS ENUM (
    'deposit',          -- 充值
    'withdrawal',       -- 提现
    'transfer_in',      -- 转入
    'transfer_out',     -- 转出
    'bonus',            -- 奖励
    'refund',           -- 退款
    'fee',              -- 手续费
    'adjustment',       -- 调整
    'freeze',           -- 冻结
    'unfreeze'          -- 解冻
);

-- 创建交易状态枚举
CREATE TYPE transaction_status AS ENUM (
    'pending',          -- 待处理
    'processing',       -- 处理中
    'completed',        -- 已完成
    'failed',           -- 失败
    'cancelled',        -- 已取消
    'expired'           -- 已过期
);

-- 创建钱包交易表
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id), -- 关联钱包
    user_id UUID NOT NULL REFERENCES users(id), -- 关联用户（冗余字段，便于查询）
    type transaction_type NOT NULL, -- 交易类型
    status transaction_status NOT NULL DEFAULT 'pending', -- 交易状态
    amount DECIMAL(20, 8) NOT NULL, -- 交易金额（TRU）
    fee DECIMAL(20, 8) NOT NULL DEFAULT 0.00, -- 手续费（TRU）
    net_amount DECIMAL(20, 8) NOT NULL, -- 净金额（amount - fee）
    balance_before DECIMAL(20, 8) NOT NULL, -- 交易前余额
    balance_after DECIMAL(20, 8) NOT NULL, -- 交易后余额
    currency_id UUID REFERENCES currencies(id), -- 涉及的货币（如果是兑换）
    exchange_rate DECIMAL(20, 8), -- 汇率（如果涉及货币兑换）
    original_amount DECIMAL(20, 8), -- 原始金额（用户输入的金额）
    reference_id VARCHAR(100), -- 外部参考ID
    reference_type VARCHAR(50), -- 参考类型：withdrawal_request, deposit, etc.
    transaction_hash VARCHAR(100), -- 交易哈希（如果适用）
    description TEXT, -- 交易描述
    metadata JSONB, -- 额外的交易数据
    processed_at TIMESTAMP WITH TIME ZONE, -- 处理时间
    processed_by UUID, -- 处理人（管理员ID，如果适用）
    expires_at TIMESTAMP WITH TIME ZONE, -- 过期时间（如果适用）
    notes TEXT, -- 备注信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束检查
    CONSTRAINT check_amount_positive CHECK (amount > 0),
    CONSTRAINT check_fee_non_negative CHECK (fee >= 0),
    CONSTRAINT check_net_amount_calculation CHECK (net_amount = amount - fee),
    CONSTRAINT check_balance_non_negative CHECK (balance_before >= 0 AND balance_after >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_id ON wallet_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_status ON wallet_transactions(status);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_amount ON wallet_transactions(amount);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_processed_at ON wallet_transactions(processed_at);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_reference ON wallet_transactions(reference_id, reference_type);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_currency ON wallet_transactions(currency_id);

-- 创建复合索引用于用户交易历史查询
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_user_history ON wallet_transactions(user_id, created_at DESC, status);

-- 创建复合索引用于钱包交易统计
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_stats ON wallet_transactions(wallet_id, type, status, created_at);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_wallet_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    
    -- 如果状态变为completed，设置处理时间
    IF NEW.status = 'completed' AND OLD.status != 'completed' AND NEW.processed_at IS NULL THEN
        NEW.processed_at = NOW();
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_wallet_transactions_updated_at
    BEFORE UPDATE ON wallet_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_wallet_transactions_updated_at();

-- 创建自动计算净金额的触发器
CREATE OR REPLACE FUNCTION calculate_net_amount()
RETURNS TRIGGER AS $$
BEGIN
    NEW.net_amount = NEW.amount - NEW.fee;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_calculate_net_amount
    BEFORE INSERT OR UPDATE ON wallet_transactions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_net_amount();