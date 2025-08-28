-- 删除触发器
DROP TRIGGER IF EXISTS trigger_create_wallet_for_new_user ON users;
DROP TRIGGER IF EXISTS trigger_wallets_updated_at ON wallets;
DROP FUNCTION IF EXISTS create_wallet_for_new_user();
DROP FUNCTION IF EXISTS update_wallets_updated_at();

-- 删除钱包表
DROP TABLE IF EXISTS wallets;

-- 删除枚举类型
DROP TYPE IF EXISTS wallet_status;