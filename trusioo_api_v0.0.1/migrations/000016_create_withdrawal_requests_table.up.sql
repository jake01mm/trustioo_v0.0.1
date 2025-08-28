-- 创建提现状态枚举
CREATE TYPE withdrawal_status AS ENUM (
    'pending',          -- 待审核
    'approved',         -- 已批准
    'processing',       -- 处理中
    'completed',        -- 已完成
    'rejected',         -- 已拒绝
    'cancelled',        -- 已取消
    'failed'            -- 失败
);

-- 创建提现申请表
CREATE TABLE IF NOT EXISTS withdrawal_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id), -- 申请用户
    wallet_id UUID NOT NULL REFERENCES wallets(id), -- 关联钱包
    bank_account_id UUID NOT NULL REFERENCES user_bank_accounts(id), -- 提现银行账户
    currency_id UUID NOT NULL REFERENCES currencies(id), -- 提现货币
    
    -- 金额信息
    amount_tru DECIMAL(20, 8) NOT NULL, -- TRU金额（从钱包扣除）
    amount_local DECIMAL(20, 8) NOT NULL, -- 本地货币金额（用户看到的）
    exchange_rate DECIMAL(20, 8) NOT NULL, -- 使用的汇率
    fee_tru DECIMAL(20, 8) NOT NULL DEFAULT 0.00, -- 手续费（TRU）
    net_amount_tru DECIMAL(20, 8) NOT NULL, -- 实际扣除的TRU金额（含手续费）
    
    -- 状态和流程
    status withdrawal_status NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 0, -- 优先级（数字越大优先级越高）
    
    -- 审核信息
    reviewed_by UUID REFERENCES admins(id), -- 审核人
    reviewed_at TIMESTAMP WITH TIME ZONE, -- 审核时间
    review_notes TEXT, -- 审核备注
    
    -- 处理信息
    processed_by UUID REFERENCES admins(id), -- 处理人
    processed_at TIMESTAMP WITH TIME ZONE, -- 处理时间
    processing_notes TEXT, -- 处理备注
    
    -- 完成信息
    completed_at TIMESTAMP WITH TIME ZONE, -- 完成时间
    transaction_reference VARCHAR(100), -- 银行转账参考号
    transaction_id UUID REFERENCES wallet_transactions(id), -- 关联的钱包交易记录
    
    -- 失败/拒绝信息
    failure_reason TEXT, -- 失败原因
    rejection_reason TEXT, -- 拒绝原因
    
    -- 用户信息（冗余，便于查询和显示）
    user_name VARCHAR(100) NOT NULL, -- 用户姓名
    user_email VARCHAR(255) NOT NULL, -- 用户邮箱
    
    -- 银行信息（冗余，便于查询和显示）
    bank_name VARCHAR(100) NOT NULL, -- 银行名称
    account_number VARCHAR(50) NOT NULL, -- 账户号码
    account_name VARCHAR(100) NOT NULL, -- 账户姓名
    
    -- 请求信息
    ip_address INET, -- 请求IP地址
    user_agent TEXT, -- 用户代理
    
    -- 过期时间
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '7 days'), -- 请求过期时间
    
    -- 元数据
    metadata JSONB, -- 额外数据
    notes TEXT, -- 备注信息
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束检查
    CONSTRAINT check_amounts_positive CHECK (
        amount_tru > 0 AND 
        amount_local > 0 AND 
        exchange_rate > 0 AND 
        fee_tru >= 0 AND 
        net_amount_tru > 0
    ),
    CONSTRAINT check_net_amount_calculation CHECK (net_amount_tru = amount_tru + fee_tru),
    CONSTRAINT check_exchange_rate_calculation CHECK (amount_local = amount_tru * exchange_rate)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_user_id ON withdrawal_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_wallet_id ON withdrawal_requests(wallet_id);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_bank_account_id ON withdrawal_requests(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_currency_id ON withdrawal_requests(currency_id);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_status ON withdrawal_requests(status);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_priority ON withdrawal_requests(priority);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_created_at ON withdrawal_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_reviewed_at ON withdrawal_requests(reviewed_at);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_processed_at ON withdrawal_requests(processed_at);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_expires_at ON withdrawal_requests(expires_at);
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_transaction_id ON withdrawal_requests(transaction_id);

-- 创建复合索引用于管理员审核队列
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_admin_queue ON withdrawal_requests(status, priority DESC, created_at ASC)
WHERE status IN ('pending', 'approved');

-- 创建复合索引用于用户提现历史
CREATE INDEX IF NOT EXISTS idx_withdrawal_requests_user_history ON withdrawal_requests(user_id, created_at DESC, status);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_withdrawal_requests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    
    -- 根据状态变化自动设置时间戳
    IF NEW.status != OLD.status THEN
        CASE NEW.status
            WHEN 'approved', 'rejected' THEN
                IF NEW.reviewed_at IS NULL THEN
                    NEW.reviewed_at = NOW();
                END IF;
            WHEN 'processing' THEN
                IF NEW.processed_at IS NULL THEN
                    NEW.processed_at = NOW();
                END IF;
            WHEN 'completed', 'failed' THEN
                IF NEW.completed_at IS NULL THEN
                    NEW.completed_at = NOW();
                END IF;
        END CASE;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_withdrawal_requests_updated_at
    BEFORE UPDATE ON withdrawal_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_withdrawal_requests_updated_at();

-- 创建自动计算净金额的触发器
CREATE OR REPLACE FUNCTION calculate_withdrawal_net_amount()
RETURNS TRIGGER AS $$
BEGIN
    NEW.net_amount_tru = NEW.amount_tru + NEW.fee_tru;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_calculate_withdrawal_net_amount
    BEFORE INSERT OR UPDATE ON withdrawal_requests
    FOR EACH ROW
    EXECUTE FUNCTION calculate_withdrawal_net_amount();