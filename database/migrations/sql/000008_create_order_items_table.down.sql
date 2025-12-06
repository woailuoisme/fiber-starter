-- 删除触发器
DROP TRIGGER IF EXISTS update_order_items_updated_at ON order_items;

-- 删除订单项表
DROP TABLE IF EXISTS order_items;
