-- 创建货币表
CREATE TABLE IF NOT EXISTS currencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(10) UNIQUE NOT NULL, -- 货币代码，如：TRU, NGN, GHS, USD
    name VARCHAR(100) NOT NULL, -- 货币名称，如：Trusioo Token, Nigerian Naira
    symbol VARCHAR(10), -- 货币符号，如：₦, ¢, $
    is_fiat BOOLEAN NOT NULL DEFAULT true, -- 是否是法定货币
    is_active BOOLEAN NOT NULL DEFAULT true, -- 是否启用
    decimal_places INTEGER NOT NULL DEFAULT 2, -- 小数位数
    display_order INTEGER NOT NULL DEFAULT 0, -- 显示顺序
    description TEXT, -- 描述信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_currencies_code ON currencies(code);
CREATE INDEX IF NOT EXISTS idx_currencies_is_active ON currencies(is_active);
CREATE INDEX IF NOT EXISTS idx_currencies_is_fiat ON currencies(is_fiat);
CREATE INDEX IF NOT EXISTS idx_currencies_display_order ON currencies(display_order);

-- 插入初始数据
INSERT INTO currencies (code, name, symbol, is_fiat, decimal_places, display_order, description) VALUES
('TRU', 'Trusioo Token', 'TRU', false, 2, 1, 'Platform token used for internal calculations'),
('NGN', 'Nigerian Naira', '₦', true, 2, 2, 'Nigerian official currency'),
('GHS', 'Ghanaian Cedi', 'GH¢', true, 2, 3, 'Ghanaian official currency'),
('USD', 'US Dollar', '$', true, 2, 4, 'United States Dollar'),
('EUR', 'Euro', '€', true, 2, 5, 'European Union currency'),
('GBP', 'British Pound', '£', true, 2, 6, 'United Kingdom currency');

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_currencies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_currencies_updated_at
    BEFORE UPDATE ON currencies
    FOR EACH ROW
    EXECUTE FUNCTION update_currencies_updated_at();