-- 删除触发器
DROP TRIGGER IF EXISTS trigger_calculate_withdrawal_net_amount ON withdrawal_requests;
DROP TRIGGER IF EXISTS trigger_withdrawal_requests_updated_at ON withdrawal_requests;
DROP FUNCTION IF EXISTS calculate_withdrawal_net_amount();
DROP FUNCTION IF EXISTS update_withdrawal_requests_updated_at();

-- 删除提现申请表
DROP TABLE IF EXISTS withdrawal_requests;

-- 删除枚举类型
DROP TYPE IF EXISTS withdrawal_status;