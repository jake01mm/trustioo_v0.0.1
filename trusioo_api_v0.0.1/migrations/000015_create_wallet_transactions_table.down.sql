-- 删除触发器
DROP TRIGGER IF EXISTS trigger_calculate_net_amount ON wallet_transactions;
DROP TRIGGER IF EXISTS trigger_wallet_transactions_updated_at ON wallet_transactions;
DROP FUNCTION IF EXISTS calculate_net_amount();
DROP FUNCTION IF EXISTS update_wallet_transactions_updated_at();

-- 删除钱包交易表
DROP TABLE IF EXISTS wallet_transactions;

-- 删除枚举类型
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;