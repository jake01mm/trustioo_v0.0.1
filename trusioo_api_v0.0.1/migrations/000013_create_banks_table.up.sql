-- 创建银行表
CREATE TABLE IF NOT EXISTS banks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL, -- 银行名称
    code VARCHAR(20) UNIQUE NOT NULL, -- 银行代码
    country_code VARCHAR(3) NOT NULL, -- 国家代码 (ISO 3166-1 alpha-3)
    currency_id UUID NOT NULL REFERENCES currencies(id), -- 主要货币
    swift_code VARCHAR(11), -- SWIFT代码
    routing_number VARCHAR(20), -- 路由号码
    is_active BOOLEAN NOT NULL DEFAULT true, -- 是否启用
    logo_url VARCHAR(500), -- 银行logo URL
    website_url VARCHAR(200), -- 银行官网
    support_phone VARCHAR(20), -- 客服电话
    support_email VARCHAR(100), -- 客服邮箱
    description TEXT, -- 描述信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_banks_code ON banks(code);
CREATE INDEX IF NOT EXISTS idx_banks_country_code ON banks(country_code);
CREATE INDEX IF NOT EXISTS idx_banks_currency_id ON banks(currency_id);
CREATE INDEX IF NOT EXISTS idx_banks_is_active ON banks(is_active);
CREATE INDEX IF NOT EXISTS idx_banks_name ON banks(name);

-- 插入尼日利亚银行数据
WITH ngn_currency AS (SELECT id FROM currencies WHERE code = 'NGN')
INSERT INTO banks (name, code, country_code, currency_id, swift_code, is_active, description)
SELECT 
    bank_data.name,
    bank_data.code,
    'NGA',
    ngn_currency.id,
    bank_data.swift_code,
    true,
    bank_data.description
FROM (VALUES
    ('Access Bank', 'ACCESS', 'ABNGNGLA', 'Access Bank Nigeria'),
    ('Guaranty Trust Bank', 'GTB', 'GTBINGLA', 'Guaranty Trust Bank Nigeria'),
    ('Zenith Bank', 'ZENITH', 'ZEIBNGLA', 'Zenith Bank Nigeria'),
    ('First Bank of Nigeria', 'FIRST', 'FBNINGLA', 'First Bank of Nigeria Limited'),
    ('United Bank for Africa', 'UBA', 'UNAFNGLA', 'United Bank for Africa'),
    ('First City Monument Bank', 'FCMB', 'FCMBNGLA', 'First City Monument Bank'),
    ('Stanbic IBTC Bank', 'STANBIC', 'SBICNGLA', 'Stanbic IBTC Bank Nigeria'),
    ('Sterling Bank', 'STERLING', 'STERNGLAX', 'Sterling Bank Nigeria'),
    ('Union Bank', 'UNION', 'UNBNNGLA', 'Union Bank of Nigeria'),
    ('Wema Bank', 'WEMA', 'WEMANGLA', 'Wema Bank Nigeria')
) AS bank_data(name, code, swift_code, description)
CROSS JOIN ngn_currency;

-- 插入加纳银行数据
WITH ghs_currency AS (SELECT id FROM currencies WHERE code = 'GHS')
INSERT INTO banks (name, code, country_code, currency_id, swift_code, is_active, description)
SELECT 
    bank_data.name,
    bank_data.code,
    'GHA',
    ghs_currency.id,
    bank_data.swift_code,
    true,
    bank_data.description
FROM (VALUES
    ('Ghana Commercial Bank', 'GCB', 'GHCBGHAC', 'Ghana Commercial Bank Limited'),
    ('Ecobank Ghana', 'ECOBANK', 'ECOCGHAC', 'Ecobank Ghana Limited'),
    ('Standard Chartered Bank Ghana', 'SCBGH', 'SCBLGHAC', 'Standard Chartered Bank Ghana'),
    ('Absa Bank Ghana', 'ABSA', 'BARCGHAC', 'Absa Bank Ghana Limited'),
    ('Fidelity Bank Ghana', 'FIDELITY', 'FBLIGHAC', 'Fidelity Bank Ghana Limited'),
    ('Zenith Bank Ghana', 'ZENITH_GH', 'ZEIBGHAC', 'Zenith Bank Ghana Limited'),
    ('CalBank Limited', 'CALBANK', 'CALBGHAC', 'CalBank Limited'),
    ('Agricultural Development Bank', 'ADB', 'ADBKGHAC', 'Agricultural Development Bank Limited'),
    ('Republic Bank Ghana', 'REPUBLIC', 'RBGHGHAC', 'Republic Bank Ghana Limited'),
    ('Stanbic Bank Ghana', 'STANBIC_GH', 'SBICGHAC', 'Stanbic Bank Ghana Limited')
) AS bank_data(name, code, swift_code, description)
CROSS JOIN ghs_currency;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_banks_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_banks_updated_at
    BEFORE UPDATE ON banks
    FOR EACH ROW
    EXECUTE FUNCTION update_banks_updated_at();