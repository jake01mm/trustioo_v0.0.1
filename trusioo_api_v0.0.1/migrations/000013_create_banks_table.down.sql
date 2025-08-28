-- 删除触发器
DROP TRIGGER IF EXISTS trigger_banks_updated_at ON banks;
DROP FUNCTION IF EXISTS update_banks_updated_at();

-- 删除银行表
DROP TABLE IF EXISTS banks;