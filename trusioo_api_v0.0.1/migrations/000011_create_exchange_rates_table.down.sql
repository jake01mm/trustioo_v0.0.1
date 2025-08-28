-- 删除触发器
DROP TRIGGER IF EXISTS trigger_exchange_rates_updated_at ON exchange_rates;
DROP FUNCTION IF EXISTS update_exchange_rates_updated_at();

-- 删除汇率表
DROP TABLE IF EXISTS exchange_rates;