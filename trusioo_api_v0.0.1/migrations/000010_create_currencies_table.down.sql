-- 删除触发器
DROP TRIGGER IF EXISTS trigger_currencies_updated_at ON currencies;
DROP FUNCTION IF EXISTS update_currencies_updated_at();

-- 删除货币表
DROP TABLE IF EXISTS currencies;