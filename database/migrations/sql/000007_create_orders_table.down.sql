-- 删除触发器
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;

-- 删除订单表
DROP TABLE IF EXISTS orders;
