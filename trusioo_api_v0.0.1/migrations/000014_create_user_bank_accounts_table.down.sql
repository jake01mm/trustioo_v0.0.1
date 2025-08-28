-- 删除触发器
DROP TRIGGER IF EXISTS trigger_ensure_single_default_account ON user_bank_accounts;
DROP TRIGGER IF EXISTS trigger_user_bank_accounts_updated_at ON user_bank_accounts;
DROP FUNCTION IF EXISTS ensure_single_default_account();
DROP FUNCTION IF EXISTS update_user_bank_accounts_updated_at();

-- 删除用户银行账户表
DROP TABLE IF EXISTS user_bank_accounts;

-- 删除枚举类型
DROP TYPE IF EXISTS bank_account_status;