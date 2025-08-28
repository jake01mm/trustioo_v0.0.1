-- 创建汇率表
CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency_id UUID NOT NULL REFERENCES currencies(id), -- 源货币（通常是TRU）
    to_currency_id UUID NOT NULL REFERENCES currencies(id), -- 目标货币
    rate DECIMAL(20, 8) NOT NULL, -- 汇率，1单位源货币 = rate单位目标货币
    is_active BOOLEAN NOT NULL DEFAULT true, -- 是否启用
    effective_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(), -- 生效时间
    effective_until TIMESTAMP WITH TIME ZONE, -- 失效时间（NULL表示无限期）
    created_by UUID, -- 创建者（管理员ID）
    notes TEXT, -- 备注信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 确保同一时间段内同一货币对只有一个活跃汇率
    CONSTRAINT check_effective_dates CHECK (effective_until IS NULL OR effective_until > effective_from),
    -- 汇率必须为正数
    CONSTRAINT check_rate_positive CHECK (rate > 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_exchange_rates_from_currency ON exchange_rates(from_currency_id);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_to_currency ON exchange_rates(to_currency_id);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_is_active ON exchange_rates(is_active);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_effective_from ON exchange_rates(effective_from);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_effective_until ON exchange_rates(effective_until);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_currency_pair ON exchange_rates(from_currency_id, to_currency_id);

-- 创建复合索引用于快速查找当前有效汇率
CREATE INDEX IF NOT EXISTS idx_exchange_rates_current_rates ON exchange_rates(from_currency_id, to_currency_id, is_active, effective_from, effective_until)
WHERE is_active = true;

-- 插入初始汇率数据（TRU到各种货币）
-- 注意：这些汇率仅为示例，实际使用时应由管理员设置
WITH tru_currency AS (SELECT id FROM currencies WHERE code = 'TRU'),
     target_currencies AS (
         SELECT id, code FROM currencies WHERE code IN ('NGN', 'GHS', 'USD', 'EUR', 'GBP')
     )
INSERT INTO exchange_rates (from_currency_id, to_currency_id, rate, notes)
SELECT 
    tru.id,
    tc.id,
    CASE 
        WHEN tc.code = 'NGN' THEN 220.00 -- 1 TRU = 220 NGN
        WHEN tc.code = 'GHS' THEN 12.50  -- 1 TRU = 12.5 GHS
        WHEN tc.code = 'USD' THEN 0.15   -- 1 TRU = 0.15 USD
        WHEN tc.code = 'EUR' THEN 0.14   -- 1 TRU = 0.14 EUR
        WHEN tc.code = 'GBP' THEN 0.12   -- 1 TRU = 0.12 GBP
    END,
    'Initial exchange rate set during system setup'
FROM tru_currency tru
CROSS JOIN target_currencies tc;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_exchange_rates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_exchange_rates_updated_at
    BEFORE UPDATE ON exchange_rates
    FOR EACH ROW
    EXECUTE FUNCTION update_exchange_rates_updated_at();