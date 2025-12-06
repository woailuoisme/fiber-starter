-- 删除触发器
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;

-- 删除支付表
DROP TABLE IF EXISTS payments;
